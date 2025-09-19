package main

import (
	"net/http"
	"ripper-backend/config"
	"ripper-backend/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	r := gin.Default()

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

	twitter := r.Group("/twitter")
	{
		twitter.POST("/account", controllers.AddTwitterAccount)
		twitter.POST("/login", controllers.TwitterLogin)
		twitter.POST("/get_tweets", controllers.GetTweets)
	}

	r.Run(":8080")
}