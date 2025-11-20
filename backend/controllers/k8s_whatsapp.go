package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ripper-backend/config"
	"ripper-backend/k8s"
	"ripper-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var k8sManager *k8s.K8sManager

// InitK8sManager initializes the Kubernetes manager
func InitK8sManager() error {
	var err error
	k8sManager, err = k8s.NewK8sManager()
	if err != nil {
		return fmt.Errorf("failed to initialize K8s manager: %v", err)
	}
	return nil
}

// GetK8sManager returns the Kubernetes manager instance
func GetK8sManager() *k8s.K8sManager {
	return k8sManager
}

// CreateWhatsAppAccount creates a new WhatsApp account by spawning a K8s pod
func CreateWhatsAppAccount(c *gin.Context) {
	// Get user ID from JWT token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	userIDStr, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// userID is already a string (UUID)
	userID := userIDStr

	// Parse request body
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Generate unique session ID
	sessionID := fmt.Sprintf("wa_%s", uuid.New().String())

	// Create database entry first
	account := models.WhatsAppAccount{
		UserID:      userID,
		SessionID:   sessionID,
		Name:        req.Name,
		Status:      "creating",
		PhoneNumber: "",
	}

	if err := config.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account record"})
		return
	}

	// If K8s manager is not available, fall back to microservice URL
	if k8sManager == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Account created (non-K8s mode)",
			"session_id":  sessionID,
			"account_id":  account.ID,
			"status":      "creating",
			"service_url": whatsappMicroserviceURL,
		})
		return
	}

	// Spawn WhatsApp pod in Kubernetes
	pod, err := k8sManager.CreateWhatsAppPod(sessionID, userID)
	if err != nil {
		// Update account status to failed
		config.DB.Model(&account).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create WhatsApp service pod",
			"details": err.Error(),
		})
		return
	}

	// Get service URL
	serviceURL := k8sManager.GetServiceURL(sessionID)

	// Update account with pod information
	config.DB.Model(&account).Updates(map[string]interface{}{
		"status":      "initializing",
		"pod_name":    pod.Name,
		"service_url": serviceURL,
	})

	// Wait for pod to be ready (async)
	go func() {
		err := k8sManager.WaitForPodReady(sessionID, 2*time.Minute)
		if err != nil {
			fmt.Printf("Pod %s failed to become ready: %v\n", pod.Name, err)
			config.DB.Model(&models.WhatsAppAccount{}).Where("session_id = ?", sessionID).
				Update("status", "failed")
			return
		}

		// Update status to ready
		config.DB.Model(&models.WhatsAppAccount{}).Where("session_id = ?", sessionID).
			Update("status", "ready")
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":     "WhatsApp service pod created successfully",
		"session_id":  sessionID,
		"account_id":  account.ID,
		"pod_name":    pod.Name,
		"service_url": serviceURL,
		"status":      "initializing",
		"note":        "Pod is being created. Use the session_id to generate QR code once ready.",
	})
}

// GetWhatsAppAccountStatus gets the status of a WhatsApp account and its pod
func GetWhatsAppAccountStatus(c *gin.Context) {
	sessionID := c.Param("session_id")

	var account models.WhatsAppAccount
	if err := config.DB.Where("session_id = ?", sessionID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// If K8s manager is available, get pod status
	if k8sManager != nil {
		podStatus, err := k8sManager.GetPodStatus(sessionID)
		if err == nil {
			account.Status = podStatus
			config.DB.Save(&account)
		}

		podIP, _ := k8sManager.GetPodIP(sessionID)

		c.JSON(http.StatusOK, gin.H{
			"session_id":  account.SessionID,
			"account_id":  account.ID,
			"name":        account.Name,
			"status":      account.Status,
			"phone":       account.PhoneNumber,
			"pod_status":  podStatus,
			"pod_ip":      podIP,
			"service_url": account.ServiceURL,
			"created_at":  account.CreatedAt,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": account.SessionID,
		"account_id": account.ID,
		"name":       account.Name,
		"status":     account.Status,
		"phone":      account.PhoneNumber,
		"created_at": account.CreatedAt,
	})
}

// GenerateQRForSession generates QR code for a specific session
func GenerateQRForSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	var account models.WhatsAppAccount
	if err := config.DB.Where("session_id = ?", sessionID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Determine the service URL
	serviceURL := account.ServiceURL
	if serviceURL == "" {
		serviceURL = whatsappMicroserviceURL
	}

	// Make request to the specific WhatsApp service pod
	resp, err := http.Post(
		fmt.Sprintf("%s/api/whatsapp/generate-qr?session_id=%s", serviceURL, sessionID),
		"application/json",
		nil,
	)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "WhatsApp service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response from WhatsApp service",
		})
		return
	}

	var qrResponse map[string]interface{}
	if err := json.Unmarshal(body, &qrResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse WhatsApp service response",
		})
		return
	}

	// Add session_id to response
	qrResponse["session_id"] = sessionID

	c.JSON(resp.StatusCode, qrResponse)
}

// DeleteWhatsAppAccountK8s deletes a WhatsApp account and its K8s pod
func DeleteWhatsAppAccountK8s(c *gin.Context) {
	accountID := c.Param("id")

	var account models.WhatsAppAccount
	if err := config.DB.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	// Delete K8s pod if manager is available
	if k8sManager != nil {
		if err := k8sManager.DeleteWhatsAppPod(account.SessionID); err != nil {
			fmt.Printf("Warning: Failed to delete K8s pod: %v\n", err)
		}
	}

	// Delete from database
	if err := config.DB.Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "WhatsApp account and pod deleted successfully",
		"session_id": account.SessionID,
	})
}

// ListWhatsAppPods lists all WhatsApp service pods in the cluster
func ListWhatsAppPods(c *gin.Context) {
	if k8sManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "K8s manager not available"})
		return
	}

	pods, err := k8sManager.ListWhatsAppPods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list pods",
			"details": err.Error(),
		})
		return
	}

	var podList []map[string]interface{}
	for _, pod := range pods {
		podList = append(podList, map[string]interface{}{
			"name":       pod.Name,
			"session_id": pod.Labels["session-id"],
			"user_id":    pod.Labels["user-id"],
			"status":     string(pod.Status.Phase),
			"pod_ip":     pod.Status.PodIP,
			"created_at": pod.CreationTimestamp,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":  podList,
		"count": len(podList),
	})
}

// GetPodLogs retrieves logs from a WhatsApp pod
func GetPodLogs(c *gin.Context) {
	sessionID := c.Param("session_id")

	if k8sManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "K8s manager not available"})
		return
	}

	logs, err := k8sManager.GetPodLogs(sessionID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get pod logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"logs":       logs,
	})
}
