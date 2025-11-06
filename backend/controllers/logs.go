package controllers

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetAPILogs godoc
// @Summary      Get API Call Logs
// @Description  Retrieve API call logs for the authenticated user
// @Tags         logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit number of logs" default(50)
// @Success      200 {array} models.ApiCallLog
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /logs [get]
func GetAPILogs(c *gin.Context) {
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

	// Get limit from query params
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	logs, err := utils.GetUserAPICallLogs(user.ID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetAPIStats godoc
// @Summary      Get API Call Statistics
// @Description  Retrieve API call statistics for the authenticated user
// @Tags         logs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /logs/stats [get]
func GetAPIStats(c *gin.Context) {
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

	stats, err := utils.GetUserAPICallStats(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
