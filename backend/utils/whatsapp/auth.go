package utils_whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	// WhatsApp server running on localhost:8083
	whatsappServerURL = "http://localhost:8083"
	httpClient        = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// QRCodeResponse represents the QR code data from WhatsApp server
type QRCodeResponse struct {
	QRCode    string `json:"qr_code"`
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

// SessionStatusResponse represents the session status from WhatsApp server
type SessionStatusResponse struct {
	SessionID   string `json:"session_id"`
	Status      string `json:"status"` // "pending", "authenticated", "failed", "expired"
	PhoneNumber string `json:"phone_number,omitempty"`
	Name        string `json:"name,omitempty"`
	Token       string `json:"token,omitempty"`
	Message     string `json:"message,omitempty"`
}

// GenerateQRCode requests a new QR code from the WhatsApp server
func GenerateQRCode() (*QRCodeResponse, error) {
	// Make request to WhatsApp server to generate QR code
	resp, err := httpClient.Post(
		fmt.Sprintf("%s/api/whatsapp/generate-qr", whatsappServerURL),
		"application/json",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WhatsApp server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WhatsApp server returned status %d: %s", resp.StatusCode, string(body))
	}

	var qrResponse QRCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&qrResponse); err != nil {
		return nil, fmt.Errorf("failed to decode QR response: %w", err)
	}

	return &qrResponse, nil
}

// CheckSessionStatus checks the authentication status of a session
func CheckSessionStatus(sessionID string) (*SessionStatusResponse, error) {
	resp, err := httpClient.Get(
		fmt.Sprintf("%s/api/whatsapp/session-status/%s", whatsappServerURL, sessionID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check session status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WhatsApp server returned status %d: %s", resp.StatusCode, string(body))
	}

	var statusResponse SessionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &statusResponse, nil
}

// DisconnectSession disconnects a WhatsApp session
func DisconnectSession(sessionID string) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/api/whatsapp/session/%s", whatsappServerURL, sessionID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create disconnect request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to disconnect session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("WhatsApp server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMessage sends a message via WhatsApp
type SendMessageRequest struct {
	SessionID   string `json:"session_id"`
	PhoneNumber string `json:"phone_number"` // Recipient phone number
	Message     string `json:"message"`
}

type SendMessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

func SendMessage(req SendMessageRequest) (*SendMessageResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := httpClient.Post(
		fmt.Sprintf("%s/api/whatsapp/send-message", whatsappServerURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	var sendResponse SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResponse); err != nil {
		return nil, fmt.Errorf("failed to decode send response: %w", err)
	}

	return &sendResponse, nil
}
