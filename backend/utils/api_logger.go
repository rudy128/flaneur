package utils

import (
	"ripper-backend/config"
	"ripper-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

// LogTwitterAPICall logs a Twitter API call to the database
func LogTwitterAPICall(c *gin.Context, userID string, twitterUsername string, endpoint string, startTime time.Time, success bool, statusCode int, errorMessage string) {
	responseTime := time.Since(startTime).Milliseconds()

	log := models.ApiCallLog{
		UserID:          userID,
		TwitterUsername: twitterUsername,
		Endpoint:        endpoint,
		Method:          c.Request.Method,
		RequestURL:      c.Request.URL.String(),
		StatusCode:      statusCode,
		Success:         success,
		ErrorMessage:    errorMessage,
		ResponseTime:    responseTime,
		IPAddress:       c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
	}

	// Save to database asynchronously to not block the response
	go func() {
		if err := config.DB.Create(&log).Error; err != nil {
			// Log error but don't fail the request
			println("Failed to log API call:", err.Error())
		}
	}()
}

// GetUserAPICallLogs retrieves API call logs for a user
func GetUserAPICallLogs(userID string, limit int) ([]models.ApiCallLog, error) {
	var logs []models.ApiCallLog

	query := config.DB.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetUserAPICallStats retrieves statistics for a user's API calls
func GetUserAPICallStats(userID string) (map[string]interface{}, error) {
	var totalCalls int64
	var successfulCalls int64
	var failedCalls int64
	var avgResponseTime float64

	// Total calls
	if err := config.DB.Model(&models.ApiCallLog{}).
		Where("user_id = ?", userID).
		Count(&totalCalls).Error; err != nil {
		return nil, err
	}

	// Successful calls
	if err := config.DB.Model(&models.ApiCallLog{}).
		Where("user_id = ? AND success = ?", userID, true).
		Count(&successfulCalls).Error; err != nil {
		return nil, err
	}

	// Failed calls
	failedCalls = totalCalls - successfulCalls

	// Average response time
	if err := config.DB.Model(&models.ApiCallLog{}).
		Where("user_id = ?", userID).
		Select("AVG(response_time)").
		Scan(&avgResponseTime).Error; err != nil {
		return nil, err
	}

	// Calls by endpoint
	var endpointStats []struct {
		Endpoint string
		Count    int64
	}
	if err := config.DB.Model(&models.ApiCallLog{}).
		Where("user_id = ?", userID).
		Select("endpoint, COUNT(*) as count").
		Group("endpoint").
		Order("count DESC").
		Scan(&endpointStats).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_calls":       totalCalls,
		"successful_calls":  successfulCalls,
		"failed_calls":      failedCalls,
		"avg_response_time": avgResponseTime,
		"calls_by_endpoint": endpointStats,
	}

	return stats, nil
}
