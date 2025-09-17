package middleware

import (
	"ripper/internal/services"
	"github.com/gorilla/websocket"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) ValidateWebSocketAuth(conn *websocket.Conn, sessionID string) (*services.SessionData, error) {
	sessionData, err := m.authService.ValidateSession(sessionID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"auth_error","message":"Invalid session"}`))
		return nil, err
	}
	return sessionData, nil
}