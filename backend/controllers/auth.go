package controllers

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/models"
	"ripper-backend/schemas"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key")

// Signup godoc
// @Summary      User Signup
// @Description  Register a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body schemas.SignupRequest true "Signup credentials"
// @Success      200 {object} schemas.MessageResponse
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Router       /auth/signup [post]
func Signup(c *gin.Context) {
	var req schemas.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	c.JSON(http.StatusOK, schemas.MessageResponse{Message: "successful"})
}

// Login godoc
// @Summary      User Login
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body schemas.LoginRequest true "Login credentials"
// @Success      200 {object} schemas.LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var req schemas.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)

	c.JSON(http.StatusOK, schemas.LoginResponse{
		Token: tokenString,
		Email: user.Email,
	})
}

// Dashboard godoc
// @Summary      Get Dashboard
// @Description  Get user dashboard information (requires authentication)
// @Tags         dashboard
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} schemas.MessageResponse
// @Failure      401 {object} map[string]string
// @Router       /dashboard [get]
func Dashboard(c *gin.Context) {
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

	c.JSON(http.StatusOK, schemas.MessageResponse{
		Message: "Hello, " + user.Name,
	})
}

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Get authenticated user's profile information (requires authentication)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /profile [get]
func GetProfile(c *gin.Context) {
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

	var twitterAccountCount int64
	config.DB.Model(&models.TwitterAccount{}).Where("user_id = ?", user.ID).Count(&twitterAccountCount)

	var whatsappAccountCount int64
	config.DB.Model(&models.WhatsAppAccount{}).Where("user_id = ?", user.ID).Count(&whatsappAccountCount)

	c.JSON(http.StatusOK, gin.H{
		"email":                   user.Email,
		"name":                    user.Name,
		"created_at":              user.CreatedAt,
		"accounts_connected":      twitterAccountCount + whatsappAccountCount,
		"twitter_accounts_count":  twitterAccountCount,
		"whatsapp_accounts_count": whatsappAccountCount,
	})
}

// ChangePassword godoc
// @Summary      Change Password
// @Description  Change authenticated user's password (requires authentication)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body schemas.ChangePasswordRequest true "Password change request"
// @Success      200 {object} schemas.MessageResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Router       /change-password [post]
func ChangePassword(c *gin.Context) {
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

	var req schemas.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err := config.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, schemas.MessageResponse{Message: "Password changed successfully"})
}
