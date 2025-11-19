package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"ripper-backend/config"
	"ripper-backend/models"

	"gorm.io/gorm"
)

// MessageScheduler handles scheduling and sending of messages
type MessageScheduler struct {
	db       *gorm.DB
	stopChan chan bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewMessageScheduler creates a new message scheduler
func NewMessageScheduler(db *gorm.DB) *MessageScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &MessageScheduler{
		db:       db,
		stopChan: make(chan bool),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the scheduler worker
func (s *MessageScheduler) Start() {
	log.Println("📅 Message Scheduler: Starting...")

	// Auto-migrate the scheduled messages table
	if err := s.db.AutoMigrate(&models.ScheduledMessage{}); err != nil {
		log.Printf("❌ Failed to migrate ScheduledMessage table: %v", err)
		return
	}

	log.Println("✅ Message Scheduler: Database migration complete")

	// Start the worker goroutine
	go s.worker()

	log.Println("✅ Message Scheduler: Worker started")
}

// Stop gracefully stops the scheduler
func (s *MessageScheduler) Stop() {
	log.Println("🛑 Message Scheduler: Stopping...")
	s.cancel()
	close(s.stopChan)
	log.Println("✅ Message Scheduler: Stopped")
}

// worker runs continuously and processes scheduled messages
func (s *MessageScheduler) worker() {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("📅 Message Scheduler: Context cancelled, stopping worker")
			return
		case <-s.stopChan:
			log.Println("📅 Message Scheduler: Stop signal received")
			return
		case <-ticker.C:
			s.processPendingMessages()
		}
	}
}

// processPendingMessages finds and sends messages that are due
func (s *MessageScheduler) processPendingMessages() {
	var messages []models.ScheduledMessage

	// Find all pending messages where scheduled_at <= now
	now := time.Now()
	result := s.db.Where("status = ? AND scheduled_at <= ?", "pending", now).
		Order("scheduled_at ASC").
		Limit(50). // Process max 50 messages at a time
		Find(&messages)

	if result.Error != nil {
		log.Printf("❌ Error fetching pending messages: %v", result.Error)
		return
	}

	if len(messages) == 0 {
		return // No messages to process
	}

	log.Printf("📨 Processing %d scheduled message(s)", len(messages))

	for _, msg := range messages {
		s.sendScheduledMessage(&msg)
	}
}

// sendScheduledMessage sends a single scheduled message
func (s *MessageScheduler) sendScheduledMessage(msg *models.ScheduledMessage) {
	log.Printf("📤 Sending scheduled message ID: %s to %s", msg.ID, msg.RecipientPhone)

	// Update status to 'sending' to prevent duplicate sends
	msg.Status = "sending"
	if err := s.db.Save(msg).Error; err != nil {
		log.Printf("❌ Failed to update message status: %v", err)
		return
	}

	// Send the message via WhatsApp microservice
	err := s.sendToWhatsApp(msg)

	now := time.Now()
	msg.SentAt = &now

	if err != nil {
		log.Printf("❌ Failed to send message %s: %v", msg.ID, err)
		msg.Status = "failed"
		msg.ErrorMessage = err.Error()
	} else {
		log.Printf("✅ Message %s sent successfully", msg.ID)
		msg.Status = "sent"
	}

	// Update the message in database
	if err := s.db.Save(msg).Error; err != nil {
		log.Printf("❌ Failed to update message after send: %v", err)
	}
}

// sendToWhatsApp sends the message to WhatsApp microservice
func (s *MessageScheduler) sendToWhatsApp(msg *models.ScheduledMessage) error {
	log.Printf("📞 Calling WhatsApp service for session: %s, phone: %s", msg.SessionID, msg.RecipientPhone)

	// Find existing message log entry (created during scheduling)
	var messageLog models.MessageLog
	result := s.db.Where("batch_id = ? AND sequence_number = ? AND status = ?", 
		msg.BatchID, msg.SequenceNumber, "pending").First(&messageLog)
	
	if result.Error != nil {
		// If no existing log found, create a new one (fallback for backward compatibility)
		log.Printf("⚠️ No existing message log found, creating new one")
		messageLog = models.MessageLog{
			UserID:         msg.UserID,
			SessionID:      msg.SessionID,
			RecipientPhone: msg.RecipientPhone,
			RecipientName:  msg.RecipientName,
			Message:        msg.Message,
			MessageType:    "scheduled",
			Status:         "pending",
			ScheduledAt:    &msg.ScheduledAt,
			BatchID:        msg.BatchID,
			SequenceNumber: msg.SequenceNumber,
			DelaySeconds:   msg.ActualDelay,
		}
		
		if err := s.db.Create(&messageLog).Error; err != nil {
			log.Printf("⚠️ Failed to create message log: %v", err)
		}
	} else {
		log.Printf("📝 Found existing message log: %s, updating status to 'sending'", messageLog.ID)
		// Update status to 'sending' to indicate we're processing it
		s.db.Model(&messageLog).Update("status", "sending")
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"session_id": msg.SessionID,
		"phone":      msg.RecipientPhone,
		"message":    msg.Message,
		"reply":      false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Update log status to failed
		if messageLog.ID != "" {
			now := time.Now()
			s.db.Model(&messageLog).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": fmt.Sprintf("Failed to marshal request: %v", err),
				"sent_at":       &now,
			})
		}
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Send to WhatsApp microservice
	baseURL := os.Getenv("WHATSAPP_MICROSERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8083"
	}
	microserviceURL := baseURL + "/api/whatsapp/send-message"

	resp, err := http.Post(
		microserviceURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		// Update log status to failed
		if messageLog.ID != "" {
			now := time.Now()
			s.db.Model(&messageLog).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": fmt.Sprintf("Microservice unavailable: %v", err),
				"sent_at":       &now,
			})
		}
		return fmt.Errorf("microservice unavailable: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// Update log status to failed
		if messageLog.ID != "" {
			now := time.Now()
			s.db.Model(&messageLog).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": fmt.Sprintf("Failed to send (status %d): %s", resp.StatusCode, string(body)),
				"sent_at":       &now,
			})
		}
		return fmt.Errorf("failed to send (status %d): %s", resp.StatusCode, string(body))
	}

	// Update log status to sent
	if messageLog.ID != "" {
		now := time.Now()
		s.db.Model(&messageLog).Updates(map[string]interface{}{
			"status":  "sent",
			"sent_at": &now,
		})
	}

	log.Printf("✅ WhatsApp message sent successfully to %s", msg.RecipientPhone)
	return nil
}

// ScheduleMessage creates a new scheduled message
func (s *MessageScheduler) ScheduleMessage(msg *models.ScheduledMessage) error {
	// Calculate actual delay if random delay is configured
	if msg.RandomDelayMin > 0 || msg.RandomDelayMax > 0 {
		min := msg.RandomDelayMin
		max := msg.RandomDelayMax

		if min > max {
			min, max = max, min
		}

		if max > 0 {
			msg.ActualDelay = min + rand.Intn(max-min+1)
		} else {
			msg.ActualDelay = min
		}

		// Add the random delay to scheduled time
		msg.ScheduledAt = msg.ScheduledAt.Add(time.Duration(msg.ActualDelay) * time.Second)

		log.Printf("🎲 Random delay applied: %d seconds (min: %d, max: %d)",
			msg.ActualDelay, msg.RandomDelayMin, msg.RandomDelayMax)
	}

	result := s.db.Create(msg)
	if result.Error != nil {
		return fmt.Errorf("failed to create scheduled message: %v", result.Error)
	}

	log.Printf("✅ Scheduled message ID: %s for %s at %s",
		msg.ID, msg.RecipientPhone, msg.ScheduledAt.Format(time.RFC3339))

	// Create MessageLog entry immediately when message is scheduled
	messageLog := &models.MessageLog{
		UserID:         msg.UserID,
		SessionID:      msg.SessionID,
		RecipientPhone: msg.RecipientPhone,
		RecipientName:  msg.RecipientName,
		Message:        msg.Message,
		MessageType:    "scheduled",
		Status:         "pending",
		ScheduledAt:    &msg.ScheduledAt,
		BatchID:        msg.BatchID,
		SequenceNumber: msg.SequenceNumber,
		DelaySeconds:   msg.ActualDelay,
	}
	
	// Save log entry to database
	if err := s.db.Create(messageLog).Error; err != nil {
		log.Printf("⚠️ Failed to create message log during scheduling: %v", err)
		// Don't fail the whole operation if logging fails
	} else {
		log.Printf("📝 Created message log entry for scheduled message: %s", messageLog.ID)
	}

	return nil
}

// ScheduleBulkMessages schedules multiple messages with delays
func (s *MessageScheduler) ScheduleBulkMessages(
	userID string,
	sessionID string,
	contacts []Contact,
	message string,
	config ScheduleConfig,
) (string, error) {

	// Generate batch ID
	batchID := fmt.Sprintf("batch_%d", time.Now().UnixNano())
	baseTime := time.Now()

	log.Printf("📦 Scheduling bulk messages: batch=%s, contacts=%d", batchID, len(contacts))

	for i, contact := range contacts {
		scheduledTime := baseTime

		if config.EnableScheduling {
			// Add cumulative delay for each message
			totalDelay := i * config.DelaySeconds
			scheduledTime = baseTime.Add(time.Duration(totalDelay) * time.Second)
		}

		// Replace {name} placeholder
		personalizedMsg := message
		if contact.Name != "" {
			personalizedMsg = replaceNamePlaceholder(message, contact.Name)
		} else {
			personalizedMsg = replaceNamePlaceholder(message, contact.Phone)
		}

		msg := &models.ScheduledMessage{
			UserID:         userID,
			SessionID:      sessionID,
			RecipientPhone: contact.Phone,
			RecipientName:  contact.Name,
			Message:        personalizedMsg,
			ScheduledAt:    scheduledTime,
			Status:         "pending",
			BatchID:        batchID,
			SequenceNumber: i + 1,
			RandomDelayMin: config.RandomDelayMin,
			RandomDelayMax: config.RandomDelayMax,
		}

		if err := s.ScheduleMessage(msg); err != nil {
			log.Printf("❌ Failed to schedule message %d: %v", i+1, err)
			continue
		}
	}

	log.Printf("✅ Bulk scheduling complete: batch=%s, total=%d", batchID, len(contacts))

	return batchID, nil
}

// ScheduleMessagesWithIndividualDelays schedules messages where each message has its own delay
func (s *MessageScheduler) ScheduleMessagesWithIndividualDelays(
	userID string,
	sessionName string,
	messages interface{},
) (string, error) {
	// Generate batch ID
	batchID := fmt.Sprintf("batch_%d", time.Now().UnixNano())
	baseTime := time.Now()

	log.Printf("📦 Scheduling messages with individual delays: batch=%s", batchID)

	// Type for messages
	type MessageWithDelay struct {
		Recipient    string `json:"recipient"`
		Message      string `json:"message"`
		DelaySeconds int    `json:"delay_seconds"`
	}

	// Convert to JSON and back to handle any struct tags
	jsonData, err := json.Marshal(messages)
	if err != nil {
		return "", fmt.Errorf("failed to marshal messages: %v", err)
	}

	var typedMessages []MessageWithDelay
	if err := json.Unmarshal(jsonData, &typedMessages); err != nil {
		return "", fmt.Errorf("failed to unmarshal messages: %v", err)
	}

	log.Printf("📨 Processing %d messages", len(typedMessages))

	for i, msg := range typedMessages {
		// Calculate scheduled time based on individual delay
		scheduledTime := baseTime.Add(time.Duration(msg.DelaySeconds) * time.Second)

		scheduledMsg := &models.ScheduledMessage{
			UserID:         userID,
			SessionID:      sessionName,
			RecipientPhone: msg.Recipient,
			Message:        msg.Message,
			ScheduledAt:    scheduledTime,
			Status:         "pending",
			BatchID:        batchID,
			SequenceNumber: i + 1,
			ActualDelay:    msg.DelaySeconds,
		}

		if err := s.ScheduleMessage(scheduledMsg); err != nil {
			log.Printf("❌ Failed to schedule message %d: %v", i+1, err)
			continue
		}

		log.Printf("✅ Scheduled message %d: delay=%ds, scheduled_at=%v", i+1, msg.DelaySeconds, scheduledTime)
	}

	log.Printf("✅ Bulk scheduling complete: batch=%s, total=%d", batchID, len(typedMessages))

	return batchID, nil
}

// Contact represents a contact with phone and name
type Contact struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

// ScheduleConfig represents scheduling configuration
type ScheduleConfig struct {
	EnableScheduling bool `json:"enable_scheduling"`
	DelaySeconds     int  `json:"delay_seconds"`
	RandomDelayMin   int  `json:"random_delay_min"`
	RandomDelayMax   int  `json:"random_delay_max"`
}

// replaceNamePlaceholder replaces {name} with actual name (case insensitive)
func replaceNamePlaceholder(message, name string) string {
	replacements := []string{"{name}", "{Name}", "{NAME}"}
	result := message

	for _, placeholder := range replacements {
		if len(result) != len(message) {
			break
		}
		result = replaceAll(result, placeholder, name)
	}

	return result
}

// replaceAll is a helper to replace all occurrences
func replaceAll(s, old, new string) string {
	result := ""
	remaining := s

	for {
		index := indexOf(remaining, old)
		if index == -1 {
			result += remaining
			break
		}
		result += remaining[:index] + new
		remaining = remaining[index+len(old):]
	}

	return result
}

// indexOf finds the first occurrence of substr in s
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// CancelScheduledMessage cancels a scheduled message
func (s *MessageScheduler) CancelScheduledMessage(messageID string, userID string) error {
	result := s.db.Model(&models.ScheduledMessage{}).
		Where("id = ? AND user_id = ? AND status = ?", messageID, userID, "pending").
		Update("status", "cancelled")

	if result.Error != nil {
		return fmt.Errorf("failed to cancel message: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("message not found or already processed")
	}

	log.Printf("✅ Cancelled scheduled message: %s", messageID)
	return nil
}

// CancelBatch cancels all pending messages in a batch
func (s *MessageScheduler) CancelBatch(batchID string, userID string) error {
	result := s.db.Model(&models.ScheduledMessage{}).
		Where("batch_id = ? AND user_id = ? AND status = ?", batchID, userID, "pending").
		Update("status", "cancelled")

	if result.Error != nil {
		return fmt.Errorf("failed to cancel batch: %v", result.Error)
	}

	log.Printf("✅ Cancelled %d messages in batch: %s", result.RowsAffected, batchID)
	return nil
}

// GetScheduledMessages retrieves scheduled messages for a user
func (s *MessageScheduler) GetScheduledMessages(userID string, status string) ([]models.ScheduledMessage, error) {
	var messages []models.ScheduledMessage

	query := s.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	result := query.Order("scheduled_at DESC").Find(&messages)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch scheduled messages: %v", result.Error)
	}

	return messages, nil
}

// GetBatchStatus retrieves the status of all messages in a batch
func (s *MessageScheduler) GetBatchStatus(batchID string, userID string) (map[string]int, error) {
	var results []struct {
		Status string
		Count  int
	}

	err := s.db.Model(&models.ScheduledMessage{}).
		Select("status, COUNT(*) as count").
		Where("batch_id = ? AND user_id = ?", batchID, userID).
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %v", err)
	}

	statusMap := make(map[string]int)
	for _, r := range results {
		statusMap[r.Status] = r.Count
	}

	return statusMap, nil
}

// Global scheduler instance
var GlobalScheduler *MessageScheduler

// InitScheduler initializes the global scheduler
func InitScheduler() {
	db := config.DB
	GlobalScheduler = NewMessageScheduler(db)
	GlobalScheduler.Start()
}

// StopScheduler stops the global scheduler
func StopScheduler() {
	if GlobalScheduler != nil {
		GlobalScheduler.Stop()
	}
}
