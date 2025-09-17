package main

import (
	"log"

	"ripper/internal/config"
	"ripper/internal/database"
	"ripper/internal/handler"
	"ripper/internal/services"
	"ripper/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	authService := services.NewAuthService(db, rdb, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)
	wsHandler := ws.NewWSHandler(authService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Session-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", handler.HealthCheck)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/ws", wsHandler.HandleWebSocket)

	r.Run(":" + cfg.Port)
}
