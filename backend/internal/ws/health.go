package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections for now
	},
}

// ServeHealthWS upgrades the HTTP connection to a WebSocket and sends health status periodically
func ServeHealthWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		msg := map[string]string{"status": "ok"}
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("WebSocket write error:", err)
			return // client disconnected
		}
		time.Sleep(1 * time.Second)
	}
}
