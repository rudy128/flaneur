package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type Message struct {
	Type    string      `json:"type"`
	Action  string      `json:"action"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

var GlobalHub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected: %s", client.userID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected: %s", client.userID)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		response := handleMessage(c, &msg)
		responseBytes, _ := json.Marshal(response)
		c.send <- responseBytes
	}
}

func HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	jwtSecret := []byte("your-secret-key")
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !parsedToken.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims := parsedToken.Claims.(jwt.MapClaims)
	email := claims["email"].(string)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		hub:    GlobalHub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: email,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func SendToUser(userID string, message Message) {
	messageBytes, _ := json.Marshal(message)
	for client := range GlobalHub.clients {
		if client.userID == userID {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(GlobalHub.clients, client)
			}
		}
	}
}