package models

type User struct {
	ID       string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"-" gorm:"not null"`
}

type TwitterAccount struct {
	ID       string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Username string `json:"username" gorm:"not null"`
	Password string `json:"password" gorm:"not null"`
	UserID   string `json:"user_id" gorm:"type:uuid;not null"`
	User     User   `json:"user" gorm:"foreignKey:UserID"`
}