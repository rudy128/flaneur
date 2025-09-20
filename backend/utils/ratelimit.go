package utils

import (
	"errors"
	"ripper-backend/config"
	"ripper-backend/models"
)

const TwitterRequestCost = 10

// CheckTwitterRateLimit checks if user has enough requests remaining
func CheckTwitterRateLimit(userID string) error {
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	if user.TwitterReqs < TwitterRequestCost {
		return errors.New("insufficient Twitter requests remaining")
	}

	return nil
}

// DeductTwitterRequests deducts requests from user's quota after successful API call
func DeductTwitterRequests(userID string) error {
	result := config.DB.Model(&models.User{}).
		Where("id = ? AND twitter_reqs >= ?", userID, TwitterRequestCost).
		Update("twitter_reqs", config.DB.Raw("twitter_reqs - ?", TwitterRequestCost))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("insufficient requests or user not found")
	}

	return nil
}

// GetTwitterRequestsRemaining returns the number of Twitter requests remaining for a user
func GetTwitterRequestsRemaining(userID string) (int, error) {
	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, errors.New("user not found")
	}

	return user.TwitterReqs, nil
}