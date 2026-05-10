package websocket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ChatHub maintains active client connections and broadcasts messages
type ChatHub struct {
	// Registered clients
	Clients map[*websocket.Conn]*Client

	// Inbound messages from clients
	Broadcast chan ChatMessage

	// Register requests from clients
	Register chan ClientConnection

	// Unregister requests from clients
	Unregister chan *websocket.Conn

	// Mutex for thread-safe access to clients
	mu sync.RWMutex

	// Chat history (store recent messages)
	History []ChatMessage
	MaxHistory int
}

// Client represents a connected WebSocket client
type Client struct {
	UserID   string
	Username string
	Conn     *websocket.Conn
	Hub      *ChatHub
	Send     chan ChatMessage
}

// NewChatHub creates a new chat hub
func NewChatHub() *ChatHub {
	return &ChatHub{
		Broadcast:  make(chan ChatMessage, 256),
		Register:   make(chan ClientConnection),
		Unregister: make(chan *websocket.Conn),
		Clients:    make(map[*websocket.Conn]*Client),
		History:    make([]ChatMessage, 0, 100),
		MaxHistory: 100,
	}
}

// Run starts the hub's main loop
func (h *ChatHub) Run() {
	for {
		select {
		case clientConn := <-h.Register:
			h.registerClient(clientConn)

		case conn := <-h.Unregister:
			h.unregisterClient(conn)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a new client to the hub
func (h *ChatHub) registerClient(clientConn ClientConnection) {
	client := &Client{
		UserID:   clientConn.UserID,
		Username: clientConn.Username,
		Conn:     clientConn.Conn,
		Hub:      h,
		Send:     make(chan ChatMessage, 256),
	}

	// Add client to hub
	h.mu.Lock()
	h.Clients[clientConn.Conn] = client
	h.mu.Unlock()

	log.Printf("Client registered: %s (%s)", client.Username, client.UserID)

	// Send join notification
	joinMsg := ChatMessage{
		UserID:    "system",
		Username:  "System",
		Message:   fmt.Sprintf("%s joined the chat", client.Username),
		Timestamp: time.Now().Unix(),
		Type:      "join",
	}

	// Send recent history to the new client
	go func() {
		h.mu.RLock()
		defer h.mu.RUnlock()
		for _, msg := range h.History {
			client.Send <- msg
		}
	}()

	// Broadcast join notification to all clients (OUTSIDE the lock)
	h.broadcastMessage(joinMsg)

	// Start the client's message reading and sending loops
	go client.readLoop()
	go client.sendLoop()
}

// unregisterClient removes a client from the hub
func (h *ChatHub) unregisterClient(conn *websocket.Conn) {
	h.mu.Lock()
	client, exists := h.Clients[conn]
	delete(h.Clients, conn)
	h.mu.Unlock()

	if !exists {
		return
	}

	close(client.Send)

	log.Printf("Client unregistered: %s (%s)", client.Username, client.UserID)

	// Send leave notification
	leaveMsg := ChatMessage{
		UserID:    "system",
		Username:  "System",
		Message:   fmt.Sprintf("%s left the chat", client.Username),
		Timestamp: time.Now().Unix(),
		Type:      "leave",
	}

	h.broadcastMessage(leaveMsg)
}

// broadcastMessage sends a message to all connected clients
func (h *ChatHub) broadcastMessage(message ChatMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Validate message
	if message.Message == "" && message.Type == "message" {
		return
	}

	// Store in history (for non-system messages)
	if message.Type == "message" {
		h.addToHistory(message)
	}

	log.Printf("Broadcasting message from %s: %s", message.Username, message.Message)

	// Send to all connected clients
	for _, client := range h.Clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			go h.closeClient(client.Conn)
		}
	}
}

// addToHistory adds a message to the chat history
func (h *ChatHub) addToHistory(message ChatMessage) {
	if len(h.History) >= h.MaxHistory {
		// Remove oldest message
		h.History = h.History[1:]
	}
	h.History = append(h.History, message)
}

// closeClient closes a client connection
func (h *ChatHub) closeClient(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Clients[conn]; exists {
		delete(h.Clients, conn)
		conn.Close()
	}
}

// GetActiveClients returns the number of active clients
func (h *ChatHub) GetActiveClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetClientList returns a list of active usernames
func (h *ChatHub) GetClientList() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	usernames := make([]string, 0, len(h.Clients))
	for _, client := range h.Clients {
		usernames = append(usernames, client.Username)
	}
	return usernames
}
