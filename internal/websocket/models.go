package websocket

import "github.com/gorilla/websocket"

// ChatMessage represents a message sent through the chat
type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "message", "join", "leave"
}

// ClientConnection represents a client connecting to the chat
type ClientConnection struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
}

// SystemMessage represents a system event (join/leave)
type SystemMessage struct {
	Type      string `json:"type"`      // "join", "leave"
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

// ErrorResponse represents an error message
type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp int64  `json:"timestamp"`
}
