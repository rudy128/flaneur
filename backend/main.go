package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ripper-backend/config"
	"ripper-backend/controllers"
	_ "ripper-backend/docs"
	"ripper-backend/scheduler"
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

	// Initialize message scheduler
	log.Println("🚀 Initializing Message Scheduler...")
	scheduler.InitScheduler()

	// Graceful shutdown handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Println("🛑 Shutting down gracefully...")
		scheduler.StopScheduler()
		os.Exit(0)
	}()

	// Start WebSocket hub
	go websocket.GlobalHub.Run()

	r := gin.Default()

	// Enable CORS for all origins
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: false, // Cannot use credentials with AllowAllOrigins
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

	whatsapp := r.Group("/whatsapp")
	{
		whatsapp.POST("/generate-qr", controllers.GenerateWhatsAppQR)
		whatsapp.GET("/session-status/:sessionId", controllers.CheckWhatsAppSession)
		whatsapp.GET("/", controllers.GetWhatsAppAccounts)
		whatsapp.POST("/send-message", controllers.SendWhatsAppMessage)
		whatsapp.POST("/send-bulk", controllers.SendBulkMessages)                  // New: Bulk send with scheduling
		whatsapp.GET("/scheduled", controllers.GetScheduledMessages)               // New: Get scheduled messages
		whatsapp.GET("/batch/:batch_id", controllers.GetBatchStatus)               // New: Get batch status
		whatsapp.DELETE("/scheduled/:message_id", controllers.CancelScheduledMessage) // New: Cancel message
		whatsapp.DELETE("/batch/:batch_id", controllers.CancelBatch)               // New: Cancel batch
	}

	logs := r.Group("/logs")
	{
		logs.GET("", controllers.GetAPILogs)
		logs.GET("/stats", controllers.GetAPIStats)
	}

	log.Println("🚀 Server starting on :8080")
	r.Run(":8080")
}
