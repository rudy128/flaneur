package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/scheduler"

	"github.com/gin-gonic/gin"
)

// SendBulkMessages sends messages to multiple contacts with optional scheduling
func SendBulkMessages(c *gin.Context) {
	var req struct {
		SessionName string `json:"session_name" binding:"required"`
		Messages    []struct {
			Recipient    string `json:"recipient" binding:"required"`
			Message      string `json:"message" binding:"required"`
			DelaySeconds int    `json:"delay_seconds"` // Individual delay for this message
		} `json:"messages" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Check if any message has a delay - if so, use scheduling
	hasDelays := false
	for _, msg := range req.Messages {
		if msg.DelaySeconds > 0 {
			hasDelays = true
			break
		}
	}

	if hasDelays {
		// Schedule messages with individual delays
		batchID, err := scheduler.GlobalScheduler.ScheduleMessagesWithIndividualDelays(
			userID,
			req.SessionName,
			req.Messages,
		)

		if err != nil {
			log.Printf("❌ Error scheduling messages: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("✅ Messages scheduled successfully: batch=%s, total=%d", batchID, len(req.Messages))
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"batch_id":  batchID,
			"scheduled": true,
			"total":     len(req.Messages),
			"message":   "Messages scheduled successfully",
		})
		return
	}

	// Send immediately without scheduling
	successCount := 0
	failCount := 0
	results := []map[string]interface{}{}
	batchID := fmt.Sprintf("batch_%d", time.Now().UnixNano())

	for i, msg := range req.Messages {
		// Create message log entry
		messageLog := &models.MessageLog{
			UserID:         userID,
			SessionID:      req.SessionName,
			RecipientPhone: msg.Recipient,
			Message:        msg.Message,
			MessageType:    "bulk",
			Status:         "pending",
			BatchID:        batchID,
			SequenceNumber: i + 1,
			DelaySeconds:   0,
		}

		// Save log entry
		if err := config.DB.Create(messageLog).Error; err != nil {
			log.Printf("⚠️ Failed to create message log: %v", err)
		}

		// Send message
		err := sendWhatsAppMessageDirect(req.SessionName, msg.Recipient, msg.Message)

		result := map[string]interface{}{
			"recipient": msg.Recipient,
			"success":   err == nil,
		}

		if err != nil {
			failCount++
			result["error"] = err.Error()

			// Update log status to failed
			if messageLog.ID != "" {
				now := time.Now()
				config.DB.Model(messageLog).Updates(map[string]interface{}{
					"status":        "failed",
					"error_message": err.Error(),
					"sent_at":       &now,
				})
			}
		} else {
			successCount++

			// Update log status to sent
			if messageLog.ID != "" {
				now := time.Now()
				config.DB.Model(messageLog).Updates(map[string]interface{}{
					"status":  "sent",
					"sent_at": &now,
				})
			}
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"scheduled":     false,
		"total":         len(req.Messages),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
	})
}

// sendImmediateBulk sends messages immediately without scheduling
func sendImmediateBulk(c *gin.Context, sessionID string, contacts []scheduler.Contact, message string) {
	successCount := 0
	failCount := 0
	results := []map[string]interface{}{}

	for _, contact := range contacts {
		// Replace {name} placeholder
		personalizedMsg := message
		if contact.Name != "" {
			personalizedMsg = replaceNamePlaceholder(message, contact.Name)
		} else {
			personalizedMsg = replaceNamePlaceholder(message, contact.Phone)
		}

		// Ensure phone has @s.whatsapp.net suffix
		phone := contact.Phone
		if !containsString(phone, "@") {
			phone = phone + "@s.whatsapp.net"
		}

		// Send message
		err := sendWhatsAppMessageDirect(sessionID, phone, personalizedMsg)

		result := map[string]interface{}{
			"phone":   contact.Phone,
			"name":    contact.Name,
			"success": err == nil,
		}

		if err != nil {
			failCount++
			result["error"] = err.Error()
		} else {
			successCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"scheduled":     false,
		"total":         len(contacts),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
	})
}

// sendWhatsAppMessageDirect sends a message directly to WhatsApp microservice
func sendWhatsAppMessageDirect(sessionID, phone, message string) error {
	jsonData, _ := json.Marshal(map[string]interface{}{
		"session_id": sessionID,
		"phone":      phone,
		"message":    message,
		"reply":      false,
	})

	// Build service URL for this session's pod
	serviceURL := fmt.Sprintf("http://whatsapp-svc-%s.%s.svc.cluster.local:8083", sessionID, k8sManager.GetNamespace())
	sendMessageURL := serviceURL + "/api/whatsapp/send-message"

	resp, err := http.Post(
		sendMessageURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("microservice unavailable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send: %s", string(body))
	}

	return nil
}

// GetScheduledMessages retrieves scheduled messages for the authenticated user
func GetScheduledMessages(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	status := c.Query("status") // Optional filter by status

	messages, err := scheduler.GlobalScheduler.GetScheduledMessages(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetBatchStatus retrieves the status of a batch
func GetBatchStatus(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	batchID := c.Param("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id required"})
		return
	}

	status, err := scheduler.GlobalScheduler.GetBatchStatus(batchID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"batch_id": batchID,
		"status":   status,
	})
}

// CancelScheduledMessage cancels a scheduled message
func CancelScheduledMessage(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	messageID := c.Param("message_id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message_id required"})
		return
	}

	err = scheduler.GlobalScheduler.CancelScheduledMessage(messageID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Scheduled message cancelled",
	})
}

// CancelBatch cancels all pending messages in a batch
func CancelBatch(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	batchID := c.Param("batch_id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_id required"})
		return
	}

	err = scheduler.GlobalScheduler.CancelBatch(batchID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Batch cancelled",
	})
}

// Helper functions

func replaceNamePlaceholder(message, name string) string {
	replacements := []string{"{name}", "{Name}", "{NAME}"}
	result := message

	for _, placeholder := range replacements {
		result = replaceAllOccurrences(result, placeholder, name)
	}

	return result
}

func replaceAllOccurrences(s, old, new string) string {
	result := ""
	remaining := s

	for {
		index := findIndex(remaining, old)
		if index == -1 {
			result += remaining
			break
		}
		result += remaining[:index] + new
		remaining = remaining[index+len(old):]
	}

	return result
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsString(s, substr string) bool {
	return findIndex(s, substr) != -1
}
