package main

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/controllers"
	_ "ripper-backend/docs"
	"ripper-backend/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Ripper Social API
// @version         1.0
// @description     API for managing Twitter data extraction and user authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@ripper-api.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	config.ConnectDB()

	// Start WebSocket hub
	go websocket.GlobalHub.Run()

	r := gin.Default()

	// Enable CORS for all origins
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Ripper API is running",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := r.Group("/auth")
	{
		auth.POST("/signup", controllers.Signup)
		auth.POST("/login", controllers.Login)
	}

	r.GET("/profile", controllers.GetProfile)
	r.POST("/change-password", controllers.ChangePassword)
	r.GET("/dashboard", controllers.Dashboard)
	r.GET("/ws", websocket.HandleWebSocket)

	twitter := r.Group("/twitter")
	{
		twitter.GET("/", controllers.GetTwitterAccounts)
		twitter.POST("/account", controllers.AddTwitterAccount)
		twitter.POST("/regenerate-token", controllers.RegenerateTwitterToken)
		twitter.POST("/post", controllers.GetTweets)
		twitter.POST("/post/likes", controllers.GetLikes)
		twitter.POST("/post/quotes", controllers.GetQuotes)
		twitter.POST("/post/comments", controllers.GetComments)
		twitter.POST("/post/reposts", controllers.GetReposts)
	}

	logs := r.Group("/logs")
	{
		logs.GET("", controllers.GetAPILogs)
		logs.GET("/stats", controllers.GetAPIStats)
	}

	r.Run(":8080")
}
