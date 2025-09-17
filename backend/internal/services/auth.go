package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ripper/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	redis     *redis.Client
	jwtSecret string
}

type SessionData struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func NewAuthService(db *gorm.DB, redis *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		redis:     redis,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(name, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, *SessionData, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	sessionID := uuid.New().String()
	sessionData := &SessionData{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	sessionJSON, _ := json.Marshal(sessionData)
	if err := s.redis.Set(context.Background(), "session:"+sessionID, sessionJSON, 24*time.Hour).Err(); err != nil {
		return "", nil, err
	}

	return sessionID, sessionData, nil
}

func (s *AuthService) ValidateSession(sessionID string) (*SessionData, error) {
	sessionJSON, err := s.redis.Get(context.Background(), "session:"+sessionID).Result()
	if err != nil {
		return nil, errors.New("invalid session")
	}

	var sessionData SessionData
	if err := json.Unmarshal([]byte(sessionJSON), &sessionData); err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func (s *AuthService) Logout(sessionID string) error {
	return s.redis.Del(context.Background(), "session:"+sessionID).Err()
}
