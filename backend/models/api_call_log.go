package models

import (
	"time"
)

// ApiCallLog represents a log entry for Twitter API calls
type ApiCallLog struct {
	ID              string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          string    `json:"user_id" gorm:"type:uuid;not null;index"`
	TwitterUsername string    `json:"twitter_username" gorm:"not null;index"`
	Endpoint        string    `json:"endpoint" gorm:"not null"`
	Method          string    `json:"method" gorm:"not null"`
	RequestURL      string    `json:"request_url" gorm:"type:text"`
	StatusCode      int       `json:"status_code"`
	Success         bool      `json:"success" gorm:"default:false"`
	ErrorMessage    string    `json:"error_message" gorm:"type:text"`
	ResponseTime    int64     `json:"response_time"` // in milliseconds
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	User            User      `json:"user" gorm:"foreignKey:UserID"`
}
