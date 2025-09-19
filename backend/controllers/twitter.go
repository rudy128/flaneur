package controllers

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/schemas"
	utils_twitter "ripper-backend/utils/twitter"
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

func GetTweets(c *gin.Context) {
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
	var req schemas.GetTweetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Try to load existing tokens
	err = utils_twitter.LoadTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found. Please login first using /twitter/login"})
		return
	}

	// Validate if the loaded session is still working
	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired. Please login again using /twitter/login"})
		return
	}

	// Extract tweet ID from URL
	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Get tweet data
	tweetWithMedia, err := utils_twitter.GetTweetWithMedia(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tweet data"})
		return
	}

	c.JSON(http.StatusOK, tweetWithMedia)
}

func TwitterLogin(c *gin.Context) {
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
	var req schemas.TwitterLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Check if login is already in progress
	if utils_twitter.IsLoginInProgress(user.ID) {
		c.JSON(http.StatusAccepted, gin.H{"message": "Login already in progress"})
		return
	}

	// Start async login
	utils_twitter.StartLoginAsync(twitterAccount.Username, twitterAccount.Password, user.ID)
	c.JSON(http.StatusAccepted, gin.H{"message": "Twitter login started in background"})
}

func GetLikes(c *gin.Context) {
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
	var req schemas.GetLikesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Try to load existing tokens
	err = utils_twitter.LoadTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found. Please login first using /twitter/login"})
		return
	}

	// Validate if the loaded session is still working
	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired. Please login again using /twitter/login"})
		return
	}

	// Extract tweet ID from URL
	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Get likers data
	likers, err := utils_twitter.GetLikers(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch likers data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"likers": likers, "count": len(likers)})
}

func GetQuotes(c *gin.Context) {
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
	var req schemas.GetQuotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Try to load existing tokens
	err = utils_twitter.LoadTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found. Please login first using /twitter/login"})
		return
	}

	// Validate if the loaded session is still working
	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired. Please login again using /twitter/login"})
		return
	}

	// Extract tweet ID from URL
	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Get quotes data
	quotes, err := utils_twitter.SearchQuotedTweets(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"quotes": quotes, "count": len(quotes)})
}

func GetComments(c *gin.Context) {
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
	var req schemas.GetCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Try to load existing tokens
	err = utils_twitter.LoadTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found. Please login first using /twitter/login"})
		return
	}

	// Validate if the loaded session is still working
	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired. Please login again using /twitter/login"})
		return
	}

	// Extract tweet ID from URL
	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Get comments data
	comments, err := utils_twitter.GetAllTweetReplies(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments, "count": len(comments)})
}

func GetReposts(c *gin.Context) {
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
	var req schemas.GetRepostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the requested username belongs to the user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Try to load existing tokens
	err = utils_twitter.LoadTokensForUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found. Please login first using /twitter/login"})
		return
	}

	// Validate if the loaded session is still working
	if !utils_twitter.ValidateSession() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired. Please login again using /twitter/login"})
		return
	}

	// Extract tweet ID from URL
	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Get reposts data
	retweeters, err := utils_twitter.GetRetweeters(tweetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reposts data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reposts": retweeters, "count": len(retweeters)})
}