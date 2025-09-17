package ws

import (
	"log"
	"net/http"

	"ripper/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

type WSResponse struct {
	Type       string                 `json:"type"`
	Data       map[string]interface{} `json:"data,omitempty"`
	StatusCode int                    `json:"status_code"`
	Message    string                 `json:"message"`
}

type WSHandler struct {
	authService *services.AuthService
}

func NewWSHandler(authService *services.AuthService) *WSHandler {
	return &WSHandler{authService: authService}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch msg.Type {
		case "register":
			h.handleRegister(conn, msg)
		case "login":
			h.handleLogin(conn, msg)
		case "logout":
			h.handleLogout(conn, msg)
		default:
			conn.WriteJSON(WSResponse{
				Type:       "error",
				StatusCode: 400,
				Message:    "Unknown message type",
			})
		}
	}
}

func (h *WSHandler) handleRegister(conn *websocket.Conn, msg WSMessage) {
	name, _ := msg.Data["name"].(string)
	email, _ := msg.Data["email"].(string)
	password, _ := msg.Data["password"].(string)
	requestId, _ := msg.Data["requestId"].(string)

	if name == "" || email == "" || password == "" {
		conn.WriteJSON(WSResponse{
			Type:       "register",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 400,
			Message:    "Name, email and password are required",
		})
		return
	}

	_, err := h.authService.Register(name, email, password)
	if err != nil {
		conn.WriteJSON(WSResponse{
			Type:       "register",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 409,
			Message:    "User already exists",
		})
		return
	}

	conn.WriteJSON(WSResponse{
		Type:       "register",
		Data:       map[string]interface{}{"requestId": requestId},
		StatusCode: 201,
		Message:    "User registered successfully",
	})
}

func (h *WSHandler) handleLogin(conn *websocket.Conn, msg WSMessage) {
	email, _ := msg.Data["email"].(string)
	password, _ := msg.Data["password"].(string)
	requestId, _ := msg.Data["requestId"].(string)

	if email == "" || password == "" {
		conn.WriteJSON(WSResponse{
			Type:       "login",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 400,
			Message:    "Email and password are required",
		})
		return
	}

	sessionID, sessionData, err := h.authService.Login(email, password)
	if err != nil {
		conn.WriteJSON(WSResponse{
			Type:       "login",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 401,
			Message:    "Invalid credentials",
		})
		return
	}

	conn.WriteJSON(WSResponse{
		Type: "login",
		Data: map[string]interface{}{
			"requestId":  requestId,
			"session_id": sessionID,
			"user": map[string]interface{}{
				"user_id":  sessionData.UserID,
				"username": sessionData.Username,
				"email":    sessionData.Email,
			},
		},
		StatusCode: 200,
		Message:    "Login successful",
	})
}

func (h *WSHandler) handleLogout(conn *websocket.Conn, msg WSMessage) {
	requestId, _ := msg.Data["requestId"].(string)

	if msg.SessionID == "" {
		conn.WriteJSON(WSResponse{
			Type:       "logout",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 400,
			Message:    "Session ID required",
		})
		return
	}

	err := h.authService.Logout(msg.SessionID)
	if err != nil {
		conn.WriteJSON(WSResponse{
			Type:       "logout",
			Data:       map[string]interface{}{"requestId": requestId},
			StatusCode: 500,
			Message:    "Logout failed",
		})
		return
	}

	conn.WriteJSON(WSResponse{
		Type:       "logout",
		Data:       map[string]interface{}{"requestId": requestId},
		StatusCode: 200,
		Message:    "Logout successful",
	})
}
