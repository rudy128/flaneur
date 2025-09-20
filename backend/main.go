package main

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/controllers"
	"ripper-backend/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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

	auth := r.Group("/auth")
	{
		auth.POST("/signup", controllers.Signup)
		auth.POST("/login", controllers.Login)
	}

	r.GET("/dashboard", controllers.Dashboard)
	r.GET("/ws", websocket.HandleWebSocket)

	twitter := r.Group("/twitter")
	{
		twitter.POST("/account", controllers.AddTwitterAccount)
		twitter.POST("/post", controllers.GetTweets)
		twitter.POST("/post/likes", controllers.GetLikes)
		twitter.POST("/post/quotes", controllers.GetQuotes)
		twitter.POST("/post/comments", controllers.GetComments)
		twitter.POST("/post/reposts", controllers.GetReposts)
	}

	r.Run(":8080")
}
