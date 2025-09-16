package handler

import (
	"ripper/internal/ws"

	"github.com/gin-gonic/gin"
)

// HealthWebSocketHandler upgrades the connection to websocket and sends health status
func HealthWebSocketHandler(c *gin.Context) {
	ws.ServeHealthWS(c.Writer, c.Request)
}
