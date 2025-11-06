package controllers

import (
	"fmt"
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/schemas"
	"ripper-backend/utils"
	utils_twitter "ripper-backend/utils/twitter"
	"strings"
	"time"

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

// Helper function to handle Twitter API calls with logging
func handleTwitterAPICall(
	c *gin.Context,
	endpoint string,
	requestHandler func(*models.TwitterAccount, schemas.GetTweetsRequest) (interface{}, error),
) {
	startTime := time.Now()

	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", endpoint, startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetTweetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	// Call the specific request handler
	result, err := requestHandler(twitterAccount, req)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, false, http.StatusInternalServerError, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, endpoint, startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, result)
}

// AddTwitterAccount godoc
// @Summary      Add Twitter Account
// @Description  Add a Twitter account for data extraction (requires JWT authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.TwitterAccountRequest true "Twitter credentials"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/account [post]
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

// GetTweets godoc
// @Summary      Get Tweet Data
// @Description  Fetch tweet data including media (requires Twitter token authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.GetTweetsRequest true "Tweet URL"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      429 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/post [post]
func GetTweets(c *gin.Context) {
	startTime := time.Now()
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", "/twitter/post", startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetTweetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	tweetWithMedia, err := utils_twitter.GetTweetWithMedia(tweetID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, false, http.StatusInternalServerError, "Failed to fetch tweet data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tweet data"})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post", startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, tweetWithMedia)
}

// GetLikes godoc
// @Summary      Get Tweet Likes
// @Description  Fetch users who liked a tweet (requires Twitter token authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.GetLikesRequest true "Tweet URL"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      429 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/post/likes [post]
func GetLikes(c *gin.Context) {
	startTime := time.Now()
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", "/twitter/post/likes", startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetLikesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	likers, err := utils_twitter.GetLikers(tweetID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, false, http.StatusInternalServerError, "Failed to fetch likers data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch likers data"})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/likes", startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, gin.H{"likers": likers, "count": len(likers)})
}

// GetQuotes godoc
// @Summary      Get Tweet Quotes
// @Description  Fetch quote tweets for a tweet (requires Twitter token authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.GetQuotesRequest true "Tweet URL"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      429 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/post/quotes [post]
func GetQuotes(c *gin.Context) {
	startTime := time.Now()
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", "/twitter/post/quotes", startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetQuotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	quotes, err := utils_twitter.SearchQuotedTweets(tweetID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, false, http.StatusInternalServerError, "Failed to fetch quotes data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quotes data"})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/quotes", startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, gin.H{"quotes": quotes, "count": len(quotes)})
}

// GetComments godoc
// @Summary      Get Tweet Comments
// @Description  Fetch comments/replies for a tweet (requires Twitter token authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.GetCommentsRequest true "Tweet URL"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      429 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/post/comments [post]
func GetComments(c *gin.Context) {
	startTime := time.Now()
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", "/twitter/post/comments", startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	comments, err := utils_twitter.GetAllTweetReplies(tweetID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, false, http.StatusInternalServerError, "Failed to fetch comments data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments data"})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/comments", startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, gin.H{"comments": comments, "count": len(comments)})
}

// GetReposts godoc
// @Summary      Get Tweet Reposts
// @Description  Fetch users who reposted a tweet (requires Twitter token authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.GetRepostsRequest true "Tweet URL"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      429 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /twitter/post/reposts [post]
func GetReposts(c *gin.Context) {
	startTime := time.Now()
	twitterAccount, err := authenticateTwitterToken(c)
	if err != nil {
		utils.LogTwitterAPICall(c, "", "", "/twitter/post/reposts", startTime, false, http.StatusUnauthorized, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	if err := utils.CheckTwitterRateLimit(twitterAccount.UserID); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusTooManyRequests, err.Error())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	var req schemas.GetRepostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusBadRequest, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = utils_twitter.LoadTokensForAccount(twitterAccount.UserID, twitterAccount.ID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusUnauthorized, "No valid Twitter session found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No valid Twitter session found"})
		return
	}

	if !utils_twitter.ValidateSession() {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusUnauthorized, "Twitter session expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Twitter session expired"})
		return
	}

	tweetID := utils_twitter.ExtractTweetID(req.URL)
	if tweetID == "" {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusBadRequest, "Invalid tweet URL")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tweet URL"})
		return
	}

	retweeters, err := utils_twitter.GetRetweeters(tweetID)
	if err != nil {
		utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, false, http.StatusInternalServerError, "Failed to fetch reposts data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reposts data"})
		return
	}

	// Deduct requests after successful API call
	utils.DeductTwitterRequests(twitterAccount.UserID)

	// Log successful call
	utils.LogTwitterAPICall(c, twitterAccount.UserID, twitterAccount.Username, "/twitter/post/reposts", startTime, true, http.StatusOK, "")

	c.JSON(http.StatusOK, gin.H{"reposts": retweeters, "count": len(retweeters)})
}

// GetTwitterAccounts godoc
// @Summary      Get Twitter Accounts
// @Description  Get all Twitter accounts connected to the authenticated user (requires JWT authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /twitter/accounts [get]
func GetTwitterAccounts(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var accounts []models.TwitterAccount
	if err := config.DB.Where("user_id = ?", user.ID).Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	response := make([]gin.H, len(accounts))
	for i, account := range accounts {
		response[i] = gin.H{
			"id":       account.ID,
			"username": account.Username,
			"token":    account.Token,
		}
	}

	c.JSON(http.StatusOK, gin.H{"accounts": response, "count": len(accounts)})
}

// RegenerateTwitterToken godoc
// @Summary      Regenerate Twitter Token
// @Description  Generate a new authentication token for a Twitter account (requires JWT authentication)
// @Tags         twitter
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.TwitterLoginRequest true "Twitter username"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /twitter/regenerate-token [post]
func RegenerateTwitterToken(c *gin.Context) {
	// Check main account authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
		return
	}

	// Parse and validate main JWT token
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
	var req schemas.RegenerateTwitterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Twitter account belongs to authenticated user
	var twitterAccount models.TwitterAccount
	if err := config.DB.Where("username = ? AND user_id = ?", req.Username, user.ID).First(&twitterAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Twitter account not found or not owned by user"})
		return
	}

	// Create new Twitter JWT token
	twitterToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"user_id":  user.ID,
	})
	twitterTokenString, _ := twitterToken.SignedString(jwtSecret)

	// Update token in database
	if err := config.DB.Model(&twitterAccount).Update("token", twitterTokenString).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   twitterTokenString,
		"message": "Twitter token regenerated successfully",
	})
}
