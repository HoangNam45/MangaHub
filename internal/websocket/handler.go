package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin properly
		// For now, allow all origins
		return true
	},
}

// JoinRequest represents a join request message
type JoinRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// HandleWebSocket handles the WebSocket connection
func (h *ChatHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}

	// Wait for join message with authentication
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	
	var joinReq JoinRequest
	err = conn.ReadJSON(&joinReq)
	
	// Clear the read deadline - readLoop will set its own
	conn.SetReadDeadline(time.Time{})
	
	if err != nil {
		log.Printf("Failed to read join request: %v", err)
		errResp := ErrorResponse{
			Error:     "Invalid join request",
			Timestamp: time.Now().Unix(),
		}
		conn.WriteJSON(errResp)
		conn.Close()
		return
	}

	// Validate user credentials
	if joinReq.UserID == "" || joinReq.Username == "" {
		log.Printf("Invalid join request: missing user_id or username")
		errResp := ErrorResponse{
			Error:     "Missing user_id or username",
			Timestamp: time.Now().Unix(),
		}
		conn.WriteJSON(errResp)
		conn.Close()
		return
	}

	// TODO: Add proper authentication here
	// For now, we accept any non-empty user_id and username
	// In production, validate JWT token or session

	// Send success message to the client
	successMsg := ChatMessage{
		UserID:    "system",
		Username:  "System",
		Message:   "Connected to chat",
		Timestamp: time.Now().Unix(),
		Type:      "system",
	}
	conn.WriteJSON(successMsg)

	// Register the client
	h.Register <- ClientConnection{
		Conn:     conn,
		UserID:   joinReq.UserID,
		Username: joinReq.Username,
	}
}

// HandleStats returns chat statistics
func (h *ChatHub) HandleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := map[string]interface{}{
		"active_clients": h.GetActiveClients(),
		"active_users":   h.GetClientList(),
		"messages_count": len(h.History),
		"timestamp":      time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(stats)
}

// HandleHistory returns recent chat history
func (h *ChatHub) HandleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.mu.RLock()
	history := make([]ChatMessage, len(h.History))
	copy(history, h.History)
	h.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": history,
	})
}
