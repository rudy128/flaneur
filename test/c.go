package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// Global client variable to be accessed by HTTP handlers
var client *whatsmeow.Client
var messageID string
var number types.JID

// Session management
type QRSession struct {
	SessionID   string
	QRCode      string
	Status      string // "pending", "authenticated", "expired", "failed"
	PhoneNumber string
	Name        string
	Client      *whatsmeow.Client
	CreatedAt   time.Time
}

var (
	sessions   = make(map[string]*QRSession)
	sessionsMu sync.RWMutex
)

// SendMessageRequest represents the JSON payload for sending messages
type SendMessageRequest struct {
	Message string `json:"message"`
	Phone   string `json:"phone"` // Phone number in format: 1234567890@s.whatsapp.net
	Reply   bool   `json:"reply"` // Whether this is a reply to a previous message
}

// SendMessageResponse represents the JSON response
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// QR Code API responses
type GenerateQRResponse struct {
	SessionID string `json:"session_id"`
	QRCode    string `json:"qr_code"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type SessionStatusResponse struct {
	SessionID   string `json:"session_id"`
	Status      string `json:"status"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Name        string `json:"name,omitempty"`
	Message     string `json:"message,omitempty"`
}

func sendmessage(message string, recipientJID types.JID, reply bool) error {
	var msg waE2E.Message
	var logMessages string

	logFile, err := os.OpenFile("event_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return err
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	logMessages += "\n<-----Send Message Log Started------>\n"

	if reply && messageID != "" {
		msg = waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(message),
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    proto.String(messageID),
					Participant: proto.String(number.String()),
				},
			},
		}
	} else {
		msg = waE2E.Message{
			Conversation: proto.String(message),
		}
	}

	if client != nil {
		sendMessage, err := client.SendMessage(context.Background(), recipientJID, &msg)
		if err != nil {
			logMessages += "Failed to send message: " + err.Error() + "\n"
			logger.Println(logMessages)
			return err
		}
		logMessages += "Message sent successfully: " + fmt.Sprintf("%+v", sendMessage) + "\n"
	} else {
		logMessages += "Client is not connected\n"
		logger.Println(logMessages)
		return fmt.Errorf("client is not connected")
	}

	logMessages += "<-----Send Message Log Ended------>\n\n"
	logger.Println(logMessages)
	return nil
}

// HTTP handler for sending WhatsApp messages
func whatsappHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Only POST method is allowed",
		})
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Invalid JSON payload: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Message field is required",
		})
		return
	}

	if req.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Phone field is required (format: 1234567890@s.whatsapp.net)",
		})
		return
	}

	// Parse JID
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Invalid phone number format: " + err.Error(),
		})
		return
	}

	// Send the message
	if err := sendmessage(req.Message, recipientJID, req.Reply); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
	})
}

// Session-aware send message handler
func sendMessageWithSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Only POST method is allowed",
		})
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		Phone     string `json:"phone"`
		Message   string `json:"message"`
		Reply     bool   `json:"reply"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Invalid JSON payload: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.SessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "session_id field is required",
		})
		return
	}

	if req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "message field is required",
		})
		return
	}

	if req.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "phone field is required (format: 1234567890@s.whatsapp.net)",
		})
		return
	}

	// Get session
	sessionsMu.RLock()
	session, exists := sessions[req.SessionID]
	sessionsMu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   fmt.Sprintf("Session %s not found", req.SessionID),
		})
		return
	}

	if session.Status != "authenticated" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   fmt.Sprintf("Session %s is not authenticated (status: %s)", req.SessionID, session.Status),
		})
		return
	}

	if session.Client == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Session client is not initialized",
		})
		return
	}

	// Parse JID
	recipientJID, err := types.ParseJID(req.Phone)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   "Invalid phone number format: " + err.Error(),
		})
		return
	}

	// Send message using the session's client
	var msg waE2E.Message
	if req.Reply && messageID != "" {
		msg = waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: proto.String(req.Message),
				ContextInfo: &waE2E.ContextInfo{
					StanzaID:    proto.String(messageID),
					Participant: proto.String(number.String()),
				},
			},
		}
	} else {
		msg = waE2E.Message{
			Conversation: proto.String(req.Message),
		}
	}

	_, err = session.Client.SendMessage(context.Background(), recipientJID, &msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to send message: %v", err),
		})
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SendMessageResponse{
		Success: true,
		Message: fmt.Sprintf("Message sent successfully via session %s", req.SessionID),
	})
}

// Generate QR code for new WhatsApp login
func generateQRHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only POST method is allowed"})
		return
	}

	// Get session_id from query parameter, or generate one if not provided
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("wa_%d", time.Now().UnixNano())
	}

	dbLog := waLog.Stdout("Database", "ERROR", true)
	ctx := context.Background()

	// Create unique DB for this session
	dbPath := fmt.Sprintf("file:session_%s.db?_foreign_keys=on", sessionID)
	container, err := sqlstore.New(ctx, "sqlite3", dbPath, dbLog)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session: " + err.Error()})
		return
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get device: " + err.Error()})
		return
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	newClient := whatsmeow.NewClient(deviceStore, clientLog)

	session := &QRSession{
		SessionID: sessionID,
		Status:    "pending",
		Client:    newClient,
		CreatedAt: time.Now(),
	}

	// Add event handler for this session
	newClient.AddEventHandler(func(evt interface{}) {
		sessionEventHandler(sessionID, evt)
	})

	// Get QR channel
	qrChan, err := newClient.GetQRChannel(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get QR channel: " + err.Error()})
		return
	}

	// Connect client
	err = newClient.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to connect: " + err.Error()})
		return
	}

	// Wait for QR code (with timeout)
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			session.QRCode = evt.Code

			sessionsMu.Lock()
			sessions[sessionID] = session
			sessionsMu.Unlock()

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GenerateQRResponse{
				SessionID: sessionID,
				QRCode:    evt.Code,
				Status:    "pending",
				Message:   "Scan QR code with WhatsApp",
			})
			return
		}
	case <-time.After(10 * time.Second):
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(map[string]string{"error": "QR code generation timeout"})
		return
	}
}

// Check session status
func sessionStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only GET method is allowed"})
		return
	}

	// Get session ID from URL path
	sessionID := r.URL.Path[len("/api/whatsapp/session-status/"):]

	sessionsMu.RLock()
	session, exists := sessions[sessionID]
	sessionsMu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Session not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SessionStatusResponse{
		SessionID:   session.SessionID,
		Status:      session.Status,
		PhoneNumber: session.PhoneNumber,
		Name:        session.Name,
		Message:     getStatusMessage(session.Status),
	})
}

func getStatusMessage(status string) string {
	switch status {
	case "pending":
		return "Waiting for QR code scan"
	case "authenticated":
		return "Successfully authenticated"
	case "expired":
		return "QR code expired"
	case "failed":
		return "Authentication failed"
	default:
		return "Unknown status"
	}
}

// Session event handler
func sessionEventHandler(sessionID string, evt interface{}) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	session, exists := sessions[sessionID]
	if !exists {
		return
	}

	switch v := evt.(type) {
	case *events.Connected:
		fmt.Printf("Session %s: Connected\n", sessionID)
		if session.Client.Store.ID != nil {
			session.Status = "authenticated"
			session.PhoneNumber = session.Client.Store.ID.User
			fmt.Printf("Session %s: Authenticated as %s\n", sessionID, session.PhoneNumber)
		}
	case *events.LoggedOut:
		fmt.Printf("Session %s: Logged out\n", sessionID)
		session.Status = "failed"
	case *events.Message:
		fmt.Printf("Session %s: Received message from %s\n", sessionID, v.Info.Sender.User)
	}
}

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
		// Store message ID and sender for potential replies
		messageID = v.Info.ID
		number = v.Info.Sender
	}
}

// loadExistingSessions loads all existing WhatsApp sessions from disk
func loadExistingSessions() error {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	clientLog := waLog.Stdout("Client", "ERROR", true)
	ctx := context.Background()

	// Find all session database files
	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Look for session database files: session_wa_*.db
		fileName := file.Name()
		if len(fileName) > 10 && fileName[:10] == "session_wa" && fileName[len(fileName)-3:] == ".db" {
			// Extract session ID from filename
			sessionID := fileName[8 : len(fileName)-3] // Remove "session_" prefix and ".db" suffix

			fmt.Printf("Loading session: %s\n", sessionID)

			// Open the database
			dbPath := fmt.Sprintf("file:%s?_foreign_keys=on", fileName)
			container, err := sqlstore.New(ctx, "sqlite3", dbPath, dbLog)
			if err != nil {
				fmt.Printf("Warning: Could not open database %s: %v\n", fileName, err)
				continue
			}

			// Get the device store
			deviceStore, err := container.GetFirstDevice(ctx)
			if err != nil {
				fmt.Printf("Warning: Could not get device from %s: %v\n", fileName, err)
				continue
			}

			// Create WhatsApp client
			waClient := whatsmeow.NewClient(deviceStore, clientLog)

			// Check if session is authenticated
			if waClient.Store.ID == nil {
				fmt.Printf("Session %s: Not authenticated, skipping\n", sessionID)
				continue
			}

			// Create session object
			session := &QRSession{
				SessionID:   sessionID,
				Status:      "authenticated",
				PhoneNumber: waClient.Store.ID.User,
				Client:      waClient,
				CreatedAt:   time.Now(),
			}

			// Add event handler for this session
			waClient.AddEventHandler(func(evt interface{}) {
				sessionEventHandler(sessionID, evt)
			})

			// Connect the client
			err = waClient.Connect()
			if err != nil {
				fmt.Printf("Warning: Could not connect session %s: %v\n", sessionID, err)
				session.Status = "failed"
			} else {
				fmt.Printf("✓ Session %s connected as %s\n", sessionID, waClient.Store.ID.User)
			}

			// Store in sessions map
			sessionsMu.Lock()
			sessions[sessionID] = session
			sessionsMu.Unlock()

			loadedCount++
		}
	}

	if loadedCount > 0 {
		fmt.Printf("\n✓ Loaded %d existing session(s) from disk\n", loadedCount)
	} else {
		fmt.Println("ℹ No existing sessions found")
	}

	return nil
}

func main() {
	// Start HTTP server first
	http.HandleFunc("/whatsapp", whatsappHandler)
	http.HandleFunc("/api/whatsapp/send-message", sendMessageWithSessionHandler) // New session-aware endpoint
	http.HandleFunc("/api/whatsapp/generate-qr", generateQRHandler)
	http.HandleFunc("/api/whatsapp/session-status/", sessionStatusHandler)

	// Add a health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		status := "disconnected"
		if client != nil && client.IsConnected() {
			status = "connected"
		}
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	})

	fmt.Println("Starting WhatsApp Microservice on :8083")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/whatsapp/generate-qr         - Generate QR code for login")
	fmt.Println("  GET  /api/whatsapp/session-status/{id} - Check session status")
	fmt.Println("  POST /api/whatsapp/send-message        - Send message via session (NEW)")
	fmt.Println("  POST /whatsapp                         - Send WhatsApp message (legacy)")
	fmt.Println("  GET  /health                           - Check connection status")

	// Start HTTP server in background
	go func() {
		if err := http.ListenAndServe(":8083", nil); err != nil {
			log.Fatal("HTTP server error:", err)
		}
	}()

	fmt.Println("\n✓ HTTP server started successfully on port 8083")
	fmt.Println("✓ Ready to accept QR generation requests")

	// Load existing sessions from disk
	fmt.Println("\n━━━ Loading Existing Sessions ━━━")
	if err := loadExistingSessions(); err != nil {
		fmt.Printf("Warning: Error loading sessions: %v\n", err)
	}

	fmt.Println("\nPress Ctrl+C to stop the service\n")

	// Optional: Initialize default client for backward compatibility
	dbLog := waLog.Stdout("Database", "ERROR", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		fmt.Println("Warning: Could not initialize default database:", err)
	} else {
		deviceStore, err := container.GetFirstDevice(ctx)
		if err != nil {
			fmt.Println("Warning: Could not get first device:", err)
		} else {
			clientLog := waLog.Stdout("Client", "ERROR", true)
			client = whatsmeow.NewClient(deviceStore, clientLog)
			client.AddEventHandler(eventHandler)

			if client.Store.ID == nil {
				fmt.Println("ℹ No saved session found. Use /api/whatsapp/generate-qr to create a new session.")
			} else {
				// Already logged in, just connect
				err = client.Connect()
				if err != nil {
					fmt.Println("Warning: Could not connect default client:", err)
				} else {
					fmt.Println("✓ Default client connected as:", client.Store.ID.User)
				}
			}
		}
	}

	// Listen to Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\n━━━ Shutting down ━━━")

	// Disconnect all sessions
	sessionsMu.RLock()
	sessionCount := len(sessions)
	sessionsMu.RUnlock()

	if sessionCount > 0 {
		fmt.Printf("Disconnecting %d session(s)...\n", sessionCount)
		sessionsMu.Lock()
		for sessionID, session := range sessions {
			if session.Client != nil && session.Client.IsConnected() {
				fmt.Printf("  Disconnecting session %s...\n", sessionID)
				session.Client.Disconnect()
			}
		}
		sessionsMu.Unlock()
	}

	// Disconnect default client
	if client != nil {
		fmt.Println("Disconnecting default client...")
		client.Disconnect()
	}

	fmt.Println("✓ All sessions disconnected")
	fmt.Println("Goodbye!")
}
