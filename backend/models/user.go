package models

import "time"

type User struct {
	ID          string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Password    string    `json:"-" gorm:"not null"`
	TwitterReqs int       `json:"twitter_reqs" gorm:"default:100"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type TwitterAccount struct {
	ID       string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username string `json:"username" gorm:"not null"`
	Password string `json:"password" gorm:"not null"`
	Token    string `json:"token" gorm:"not null"`
	UserID   string `json:"user_id" gorm:"type:uuid;not null"`
	User     User   `json:"user" gorm:"foreignKey:UserID"`
}

type WhatsAppAccount struct {
	ID          string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PhoneNumber string    `json:"phone_number" gorm:"not null"`
	Name        string    `json:"name"`
	SessionID   string    `json:"session_id" gorm:"not null;unique"`
	Status      string    `json:"status" gorm:"default:'active'"` // active, disconnected
	UserID      string    `json:"user_id" gorm:"type:uuid;not null"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
