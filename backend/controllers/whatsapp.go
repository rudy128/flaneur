package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ripper-backend/config"
	"ripper-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const whatsappMicroserviceURL = "http://localhost:8083"

// Helper function to extract user ID from JWT token
func getUserIDFromToken(tokenString string) (string, error) {
	fmt.Println("DEBUG: Parsing JWT token...")
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		fmt.Println("DEBUG: JWT parse error:", err)
		return "", fmt.Errorf("invalid token: %v", err)
	}
	
	if !token.Valid {
		fmt.Println("DEBUG: Token is not valid")
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("DEBUG: Failed to parse claims")
		return "", fmt.Errorf("invalid token claims")
	}
	
	fmt.Println("DEBUG: Token claims:", claims)
	
	email, ok := claims["email"].(string)
	if !ok {
		fmt.Println("DEBUG: Email not found in claims")
		return "", fmt.Errorf("email not found in token")
	}
	
	fmt.Println("DEBUG: Email from token:", email)

	// Get user ID from email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		fmt.Println("DEBUG: User lookup error:", err)
		return "", fmt.Errorf("user not found")
	}
	
	fmt.Println("DEBUG: User found, ID:", user.ID)

	return user.ID, nil
}

// WhatsAppAccount represents a connected WhatsApp account
type WhatsAppAccount struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name,omitempty"`
	SessionID   string    `json:"session_id"`
	Status      string    `json:"status"`
	ConnectedAt time.Time `json:"connected_at"`
}

// GenerateWhatsAppQR initiates QR code generation for WhatsApp login
func GenerateWhatsAppQR(c *gin.Context) {
	// Forward request to WhatsApp microservice
	resp, err := http.Post(
		whatsappMicroserviceURL+"/api/whatsapp/generate-qr",
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

	// Forward the response
	var qrResponse map[string]interface{}
	if err := json.Unmarshal(body, &qrResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse WhatsApp service response",
		})
		return
	}

	c.JSON(resp.StatusCode, qrResponse)
}

// CheckWhatsAppSession checks the status of a WhatsApp login session
func CheckWhatsAppSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	// Forward request to WhatsApp microservice
	resp, err := http.Get(
		fmt.Sprintf("%s/api/whatsapp/session-status/%s", whatsappMicroserviceURL, sessionID),
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

	var statusResponse map[string]interface{}
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse WhatsApp service response",
		})
		return
	}

	// If authenticated, save to database
	if status, ok := statusResponse["status"].(string); ok && status == "authenticated" {
		// Get user ID from JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := getUserIDFromToken(token)
			if err == nil {
				// Save WhatsApp account to database
				phoneNumber, _ := statusResponse["phone_number"].(string)
				name, _ := statusResponse["name"].(string)

				whatsappAccount := models.WhatsAppAccount{
					PhoneNumber: phoneNumber,
					Name:        name,
					SessionID:   sessionID,
					Status:      "active",
					UserID:      userID,
				}

				// Check if account already exists
				var existingAccount models.WhatsAppAccount
				result := config.DB.Where("session_id = ?", sessionID).First(&existingAccount)

				if result.Error != nil {
					// Create new account
					config.DB.Create(&whatsappAccount)
				} else {
					// Update existing account
					config.DB.Model(&existingAccount).Updates(whatsappAccount)
				}
			}
		}
	}

	c.JSON(resp.StatusCode, statusResponse)
}

// GetWhatsAppAccounts returns all connected WhatsApp accounts for the user
func GetWhatsAppAccounts(c *gin.Context) {
	// Get user ID from JWT token
	authHeader := c.GetHeader("Authorization")
	fmt.Println("DEBUG: Authorization header:", authHeader)
	
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		fmt.Println("DEBUG: Missing or invalid authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	tokenPreview := token
	if len(token) > 20 {
		tokenPreview = token[:20] + "..."
	}
	fmt.Println("DEBUG: Token extracted (length:", len(token), "):", tokenPreview)
	
	userID, err := getUserIDFromToken(token)
	if err != nil {
		fmt.Println("DEBUG: getUserIDFromToken error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
			"details": err.Error(), // Add details for debugging
		})
		return
	}
	
	fmt.Println("DEBUG: User ID from token:", userID)

	// Fetch WhatsApp accounts from database
	var accounts []models.WhatsAppAccount
	if err := config.DB.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		fmt.Println("DEBUG: Database error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch WhatsApp accounts"})
		return
	}
	
	fmt.Println("DEBUG: Found accounts:", len(accounts))

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"count":    len(accounts),
	})
}

// SendWhatsAppMessage sends a message via WhatsApp
func SendWhatsAppMessage(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
		Message   string `json:"message" binding:"required"`
		Reply     bool   `json:"reply"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("DEBUG SendMessage: Invalid JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("DEBUG SendMessage: Received request - SessionID: %s, Phone: %s, Message: %s\n", 
		req.SessionID, req.Phone, req.Message)

	// Forward to WhatsApp microservice with session_id
	jsonData, _ := json.Marshal(map[string]interface{}{
		"session_id": req.SessionID,
		"phone":      req.Phone,
		"message":    req.Message,
		"reply":      req.Reply,
	})

	microserviceURL := whatsappMicroserviceURL + "/api/whatsapp/send-message"
	fmt.Printf("DEBUG SendMessage: Forwarding to microservice: %s\n", microserviceURL)
	fmt.Printf("DEBUG SendMessage: Payload: %s\n", string(jsonData))

	resp, err := http.Post(
		microserviceURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("DEBUG SendMessage: Microservice error: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "WhatsApp service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("DEBUG SendMessage: Failed to read response: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response from WhatsApp service",
		})
		return
	}

	fmt.Printf("DEBUG SendMessage: Microservice response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG SendMessage: Microservice response body: %s\n", string(body))

	var sendResponse map[string]interface{}
	if err := json.Unmarshal(body, &sendResponse); err != nil {
		fmt.Printf("DEBUG SendMessage: Failed to parse response: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse WhatsApp service response",
		})
		return
	}

	c.JSON(resp.StatusCode, sendResponse)
}
