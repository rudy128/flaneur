package models

import "time"

// MessageLog represents a detailed log of every message sent through the system
type MessageLog struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         string     `json:"user_id" gorm:"type:uuid;not null;index"`
	SessionID      string     `json:"session_id" gorm:"not null;index"`
	RecipientPhone string     `json:"recipient_phone" gorm:"not null;index"`
	RecipientName  string     `json:"recipient_name"`
	Message        string     `json:"message" gorm:"type:text;not null"`
	MessageType    string     `json:"message_type" gorm:"default:'bulk'"` // bulk, single, scheduled
	Status         string     `json:"status" gorm:"default:'pending';index"` // pending, sent, failed
	ScheduledAt    *time.Time `json:"scheduled_at"`
	SentAt         *time.Time `json:"sent_at" gorm:"index"`
	ErrorMessage   string     `json:"error_message"`
	BatchID        string     `json:"batch_id" gorm:"index"` // Group messages from same bulk send
	SequenceNumber int        `json:"sequence_number"`
	DelaySeconds   int        `json:"delay_seconds"` // Delay before sending (cumulative)
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	User           User       `json:"user" gorm:"foreignKey:UserID"`
}
