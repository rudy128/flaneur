package controllers

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/schemas"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AddTwitterAccount(c *gin.Context) {
	// Check authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
		return
	}

	// Parse and validate token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Extract email from token
	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	// Find user by email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Bind request body
	var req schemas.TwitterAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create Twitter account
	twitterAccount := models.TwitterAccount{
		Username: req.Username,
		Password: req.Password,
		UserID:   user.ID,
	}

	if err := config.DB.Create(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Twitter account"})
		return
	}

	// Return the created account data
	c.JSON(http.StatusOK, twitterAccount)
}