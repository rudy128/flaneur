package main

import (
	"ripper/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", handler.HealthWebSocketHandler)

	r.Run(":8080")
}
