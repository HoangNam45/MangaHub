package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// ProgressUpdate represents a reading progress update
type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
}

// AuthMessage represents client authentication
type AuthMessage struct {
	Type     string `json:"type"` // "auth"
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// ServerMessage represents messages sent from server to client
type ServerMessage struct {
	Type    string          `json:"type"` // "auth_response", "progress_update", "error"
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ClientConnection represents an active client connection
type ClientConnection struct {
	UserID   string
	Username string
	Conn     net.Conn
	Done     chan bool
}

// ProgressSyncServer is the TCP server for progress synchronization
type ProgressSyncServer struct {
	Port           string
	Listener       net.Listener
	Connections    map[string][]*ClientConnection // userID -> list of connections
	ConnectionsMu  sync.RWMutex
	Broadcast      chan ProgressUpdate
	Done           chan bool
	MaxConnections int
}

// NewProgressSyncServer creates a new TCP server instance
func NewProgressSyncServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		Port:           port,
		Connections:    make(map[string][]*ClientConnection),
		Broadcast:      make(chan ProgressUpdate, 100),
		Done:           make(chan bool),
		MaxConnections: 1000,
	}
}

// Start starts the TCP server
func (ps *ProgressSyncServer) Start() error {
	listener, err := net.Listen("tcp", ":"+ps.Port)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}

	ps.Listener = listener
	fmt.Printf("TCP Progress Sync Server started on port %s\n", ps.Port)

	// Handle broadcasts in a goroutine
	go ps.handleBroadcasts()

	// Accept connections in a goroutine
	go ps.acceptConnections()

	return nil
}

// acceptConnections accepts incoming TCP connections
func (ps *ProgressSyncServer) acceptConnections() {
	for {
		select {
		case <-ps.Done:
			return
		default:
			conn, err := ps.Listener.Accept()
			if err != nil {
				continue
			}

			// Check if at capacity
			ps.ConnectionsMu.RLock()
			totalConnections := 0
			for _, conns := range ps.Connections {
				totalConnections += len(conns)
			}
			ps.ConnectionsMu.RUnlock()

			if totalConnections >= ps.MaxConnections {
				response := ServerMessage{
					Type:    "error",
					Status:  "failed",
					Message: "Server at capacity",
				}
				respBytes, _ := json.Marshal(response)
				conn.Write(respBytes)
				conn.Close()
				continue
			}

			// Handle new connection
			go ps.handleConnection(conn)
		}
	}
}

// handleConnection handles a new client connection
func (ps *ProgressSyncServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set read deadline for authentication
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read authentication message
	decoder := json.NewDecoder(conn)
	var authMsg AuthMessage

	err := decoder.Decode(&authMsg)
	if err != nil {
		response := ServerMessage{
			Type:    "error",
			Status:  "failed",
			Message: "Authentication failed: invalid message format",
		}
		respBytes, _ := json.Marshal(response)
		conn.Write(respBytes)
		return
	}

	// Validate authentication
	if authMsg.Type != "auth" || authMsg.UserID == "" {
		response := ServerMessage{
			Type:    "error",
			Status:  "failed",
			Message: "Authentication failed: missing credentials",
		}
		respBytes, _ := json.Marshal(response)
		conn.Write(respBytes)
		return
	}

	// TODO: Validate token with JWT service in production
	// For now, accept all authenticated requests

	// Clear read deadline after authentication
	conn.SetReadDeadline(time.Time{})

	// Create client connection object
	clientConn := &ClientConnection{
		UserID:   authMsg.UserID,
		Username: authMsg.Username,
		Conn:     conn,
		Done:     make(chan bool),
	}

	// Add to active connections
	ps.ConnectionsMu.Lock()
	ps.Connections[authMsg.UserID] = append(ps.Connections[authMsg.UserID], clientConn)
	ps.ConnectionsMu.Unlock()

	// Send authentication success
	response := ServerMessage{
		Type:    "auth_response",
		Status:  "success",
		Message: fmt.Sprintf("Welcome %s, connection established", authMsg.Username),
	}
	respBytes, _ := json.Marshal(response)
	conn.Write(respBytes)

	fmt.Printf("Client connected: %s (UserID: %s)\n", authMsg.Username, authMsg.UserID)

	// Keep connection alive and listen for messages
	ps.listenToClient(clientConn)

	// Remove from active connections on disconnect
	ps.removeConnection(authMsg.UserID, clientConn)
	fmt.Printf("Client disconnected: %s (UserID: %s)\n", authMsg.Username, authMsg.UserID)
}

// listenToClient listens to messages from a connected client
func (ps *ProgressSyncServer) listenToClient(clientConn *ClientConnection) {
	decoder := json.NewDecoder(clientConn.Conn)

	for {
		select {
		case <-clientConn.Done:
			return
		default:
			var msg map[string]interface{}
			err := decoder.Decode(&msg)
			if err != nil {
				return
			}

			// Handle different message types
			if msgType, ok := msg["type"].(string); ok {
				if msgType == "ping" {
					// Respond to ping
					response := ServerMessage{
						Type:    "pong",
						Status:  "success",
						Message: "pong",
					}
					respBytes, _ := json.Marshal(response)
					clientConn.Conn.Write(respBytes)
				}
			}
		}
	}
}

// handleBroadcasts handles broadcasting progress updates to clients
func (ps *ProgressSyncServer) handleBroadcasts() {
	for {
		select {
		case <-ps.Done:
			return
		case update := <-ps.Broadcast:
			ps.broadcastToUser(update.UserID, update)
		}
	}
}

// broadcastToUser sends a progress update to all connections of a specific user
func (ps *ProgressSyncServer) broadcastToUser(userID string, update ProgressUpdate) {
	ps.ConnectionsMu.RLock()
	connections, exists := ps.Connections[userID]
	ps.ConnectionsMu.RUnlock()

	if !exists || len(connections) == 0 {
		return
	}

	// Create server message
	response := ServerMessage{
		Type:   "progress_update",
		Status: "success",
	}

	// Marshal the update as payload
	payloadBytes, err := json.Marshal(update)
	if err != nil {
		fmt.Printf("Failed to marshal update: %v\n", err)
		return
	}
	response.Payload = payloadBytes

	respBytes, _ := json.Marshal(response)

	// Send to all connections of this user
	var failedConnections []*ClientConnection
	for _, conn := range connections {
		_, err := conn.Conn.Write(respBytes)
		if err != nil {
			failedConnections = append(failedConnections, conn)
		}
	}

	// Remove failed connections
	if len(failedConnections) > 0 {
		ps.ConnectionsMu.Lock()
		for _, failedConn := range failedConnections {
			ps.removeConnectionUnsafe(userID, failedConn)
		}
		ps.ConnectionsMu.Unlock()
	}
}

// removeConnection removes a connection from the active list
func (ps *ProgressSyncServer) removeConnection(userID string, clientConn *ClientConnection) {
	ps.ConnectionsMu.Lock()
	defer ps.ConnectionsMu.Unlock()
	ps.removeConnectionUnsafe(userID, clientConn)
}

// removeConnectionUnsafe removes a connection without locking (must be called with lock held)
func (ps *ProgressSyncServer) removeConnectionUnsafe(userID string, clientConn *ClientConnection) {
	connections, exists := ps.Connections[userID]
	if !exists {
		return
	}

	for i, conn := range connections {
		if conn == clientConn {
			ps.Connections[userID] = append(connections[:i], connections[i+1:]...)
			clientConn.Conn.Close()
			close(clientConn.Done)

			if len(ps.Connections[userID]) == 0 {
				delete(ps.Connections, userID)
			}
			break
		}
	}
}

// BroadcastUpdate sends a progress update to connected clients
func (ps *ProgressSyncServer) BroadcastUpdate(update ProgressUpdate) {
	select {
	case ps.Broadcast <- update:
	default:
		fmt.Println("Broadcast channel full, update dropped")
	}
}

// Stop gracefully shuts down the server
func (ps *ProgressSyncServer) Stop() {
	fmt.Println("Shutting down TCP server...")

	// Close all connections
	ps.ConnectionsMu.Lock()
	for userID, connections := range ps.Connections {
		for _, conn := range connections {
			conn.Conn.Close()
			close(conn.Done)
		}
		delete(ps.Connections, userID)
	}
	ps.ConnectionsMu.Unlock()

	// Close listener
	if ps.Listener != nil {
		ps.Listener.Close()
	}

	close(ps.Done)
	fmt.Println("TCP server stopped")
}

// GetActiveConnections returns the number of active connections
func (ps *ProgressSyncServer) GetActiveConnections() int {
	ps.ConnectionsMu.RLock()
	defer ps.ConnectionsMu.RUnlock()

	total := 0
	for _, conns := range ps.Connections {
		total += len(conns)
	}
	return total
}

// GetConnectionsForUser returns the number of connections for a specific user
func (ps *ProgressSyncServer) GetConnectionsForUser(userID string) int {
	ps.ConnectionsMu.RLock()
	defer ps.ConnectionsMu.RUnlock()

	connections, exists := ps.Connections[userID]
	if !exists {
		return 0
	}
	return len(connections)
}
