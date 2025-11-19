package controllers

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMessageLogs retrieves message history for the authenticated user
func GetMessageLogs(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get query parameters
	status := c.Query("status")       // Optional filter by status (pending, sent, failed)
	batchID := c.Query("batch_id")    // Optional filter by batch
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Build query
	query := config.DB.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}

	// Get total count
	var total int64
	query.Model(&models.MessageLog{}).Count(&total)

	// Get logs with pagination, ordered by most recent first
	var logs []models.MessageLog
	err = query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve message logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetMessageLogStats retrieves statistics about message logs
func GetMessageLogStats(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := c.GetHeader("Authorization")[len("Bearer "):]
	userID, err := getUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get counts by status
	var stats struct {
		Total   int64 `json:"total"`
		Pending int64 `json:"pending"`
		Sent    int64 `json:"sent"`
		Failed  int64 `json:"failed"`
	}

	config.DB.Model(&models.MessageLog{}).Where("user_id = ?", userID).Count(&stats.Total)
	config.DB.Model(&models.MessageLog{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&stats.Pending)
	config.DB.Model(&models.MessageLog{}).Where("user_id = ? AND status = ?", userID, "sent").Count(&stats.Sent)
	config.DB.Model(&models.MessageLog{}).Where("user_id = ? AND status = ?", userID, "failed").Count(&stats.Failed)

	c.JSON(http.StatusOK, stats)
}
