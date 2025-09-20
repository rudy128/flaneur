package controllers

import (
	"fmt"
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/schemas"
	utils_twitter "ripper-backend/utils/twitter"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func authenticateTwitterToken(c *gin.Context) (*models.TwitterAccount, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("missing or invalid token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("token = ?", tokenString).First(&twitterAccount).Error; err != nil {
		return nil, fmt.Errorf("invalid Twitter token")
	}

	return &twitterAccount, nil
}

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

	// Create Twitter JWT token (infinite duration)
	twitterToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"user_id":  user.ID,
	})
	twitterTokenString, _ := twitterToken.SignedString(jwtSecret)

	// Create Twitter account
	twitterAccount := models.TwitterAccount{
		Username: req.Username,
		Password: req.Password,
		Token:    twitterTokenString,
		UserID:   user.ID,
	}

	if err := config.DB.Create(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Twitter account"})
		return
	}

	// Start Twitter login in background
	utils_twitter.StartLoginAsync(twitterAccount.Username, twitterAccount.Password, user.ID, twitterAccount.ID)

	// Return the created account data with token
	c.JSON(http.StatusOK, gin.H{
		"id":       twitterAccount.ID,
		"username": twitterAccount.Username,
		"token":    twitterTokenString,
		"user_id":  twitterAccount.UserID,
		"message":  "Twitter account created and login started in background",
	})
}

func GetTweets(c *gin.Context) {
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetTweetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	tweetWithMedia, err := utils_twitter.GetTweetWithMedia(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tweet data"})
		return
	}

	c.JSON(http.StatusOK, tweetWithMedia)
}

func GetLikes(c *gin.Context) {
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetLikesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	likers, err := utils_twitter.GetLikers(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch likers data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"likers": likers, "count": len(likers)})
}

func GetQuotes(c *gin.Context) {
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetQuotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	quotes, err := utils_twitter.SearchQuotedTweets(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quotes": quotes, "count": len(quotes)})
}

func GetComments(c *gin.Context) {
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	comments, err := utils_twitter.GetAllTweetReplies(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments, "count": len(comments)})
}

func GetReposts(c *gin.Context) {
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetRepostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	retweeters, err := utils_twitter.GetRetweeters(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reposts data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reposts": retweeters, "count": len(retweeters)})
}
