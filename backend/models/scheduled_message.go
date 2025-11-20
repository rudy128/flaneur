package models

import "time"

// ScheduledMessage represents a message scheduled to be sent at a specific time
type ScheduledMessage struct {
	ID             string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID         string     `json:"user_id" gorm:"type:uuid;not null;index"`
	SessionID      string     `json:"session_id" gorm:"not null"`
	RecipientPhone string     `json:"recipient_phone" gorm:"not null"`
	RecipientName  string     `json:"recipient_name"`
	Message        string     `json:"message" gorm:"type:text;not null"`
	ScheduledAt    time.Time  `json:"scheduled_at" gorm:"not null;index"`
	Status         string     `json:"status" gorm:"default:'pending';index"` // pending, sent, failed, cancelled, paused
	SentAt         *time.Time `json:"sent_at"`
	ErrorMessage   string     `json:"error_message"`
	BatchID        string     `json:"batch_id" gorm:"index"` // Group messages from same bulk send
	SequenceNumber int        `json:"sequence_number"`       // Order within batch
	RandomDelayMin int        `json:"random_delay_min"`      // Min delay in seconds
	RandomDelayMax int        `json:"random_delay_max"`      // Max delay in seconds
	ActualDelay    int        `json:"actual_delay"`          // Actual delay used in seconds
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	User           User       `json:"user" gorm:"foreignKey:UserID"`
}

// ScheduleConfig represents the scheduling configuration for a batch
type ScheduleConfig struct {
	EnableScheduling bool `json:"enable_scheduling"` // Whether to enable scheduling
	DelaySeconds     int  `json:"delay_seconds"`     // Fixed delay between messages (0 = immediate)
	RandomDelayMin   int  `json:"random_delay_min"`  // Minimum random delay in seconds
	RandomDelayMax   int  `json:"random_delay_max"`  // Maximum random delay in seconds
}
