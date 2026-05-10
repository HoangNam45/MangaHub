package udp

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Notification represents a chapter release notification
type Notification struct {
	Type      string `json:"type"` // "chapter_release", "manga_update"
	MangaID   string `json:"manga_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// ClientRegistration represents a client registration request
type ClientRegistration struct {
	Type       string `json:"type"` // "register"
	UserID     string `json:"user_id"`
	Preferences map[string]interface{} `json:"preferences,omitempty"` // e.g., {"genres": ["action", "adventure"]}
}

// RegistrationResponse is sent back to client after registration
type RegistrationResponse struct {
	Type    string `json:"type"` // "registration_confirmed"
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RegisteredClient represents a registered UDP client
type RegisteredClient struct {
	UserID      string
	Address     *net.UDPAddr
	Registered  time.Time
	Preferences map[string]interface{}
}

// NotificationServer is a UDP server for broadcasting notifications
type NotificationServer struct {
	Port        string
	Conn        *net.UDPConn
	Clients     map[string]*RegisteredClient // userID -> client
	ClientsMu   sync.RWMutex
	Broadcast   chan Notification
	Register    chan *RegisteredClient
	Unregister  chan string
	Done        chan bool
	MaxClients  int
}

// NewNotificationServer creates a new UDP notification server
func NewNotificationServer(port string) *NotificationServer {
	return &NotificationServer{
		Port:       port,
		Clients:    make(map[string]*RegisteredClient),
		Broadcast:  make(chan Notification, 100),
		Register:   make(chan *RegisteredClient, 10),
		Unregister: make(chan string, 10),
		Done:       make(chan bool),
		MaxClients: 5000,
	}
}

// Start starts the UDP notification server
func (ns *NotificationServer) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+ns.Port)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %v", err)
	}

	ns.Conn = conn
	fmt.Printf("UDP Notification Server started on port %s\n", ns.Port)

	// Start message handlers
	go ns.handleMessages()
	go ns.handleBroadcasts()
	go ns.listenForRegistrations()

	return nil
}

// listenForRegistrations listens for incoming UDP registration packets
func (ns *NotificationServer) listenForRegistrations() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ns.Done:
			return
		default:
			n, remoteAddr, err := ns.Conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("Error reading UDP: %v\n", err)
				continue
			}

			// Parse registration message
			var registration ClientRegistration
			err = json.Unmarshal(buffer[:n], &registration)
			if err != nil {
				fmt.Printf("Invalid registration message: %v\n", err)
				continue
			}

			if registration.Type == "register" {
				ns.handleRegistration(&registration, remoteAddr)
			}
		}
	}
}

// handleRegistration processes a client registration request
func (ns *NotificationServer) handleRegistration(reg *ClientRegistration, remoteAddr *net.UDPAddr) {
	// Check if at capacity
	ns.ClientsMu.RLock()
	numClients := len(ns.Clients)
	ns.ClientsMu.RUnlock()

	if numClients >= ns.MaxClients {
		response := RegistrationResponse{
			Type:    "registration_failed",
			Status:  "failed",
			Message: "Server at capacity",
		}
		respBytes, _ := json.Marshal(response)
		ns.Conn.WriteToUDP(respBytes, remoteAddr)
		return
	}

	// Create registered client
	client := &RegisteredClient{
		UserID:      reg.UserID,
		Address:     remoteAddr,
		Registered:  time.Now(),
		Preferences: reg.Preferences,
	}

	// Add to clients
	ns.ClientsMu.Lock()
	ns.Clients[reg.UserID] = client
	ns.ClientsMu.Unlock()

	// Send confirmation
	response := RegistrationResponse{
		Type:    "registration_confirmed",
		Status:  "success",
		Message: fmt.Sprintf("Successfully registered for notifications, UserID: %s", reg.UserID),
	}
	respBytes, _ := json.Marshal(response)
	ns.Conn.WriteToUDP(respBytes, remoteAddr)

	fmt.Printf("Client registered: UserID %s from %s\n", reg.UserID, remoteAddr.String())
}

// handleMessages processes channel messages
func (ns *NotificationServer) handleMessages() {
	for {
		select {
		case <-ns.Done:
			return

		case client := <-ns.Register:
			ns.ClientsMu.Lock()
			ns.Clients[client.UserID] = client
			ns.ClientsMu.Unlock()
			fmt.Printf("Client added: %s\n", client.UserID)

		case userID := <-ns.Unregister:
			ns.ClientsMu.Lock()
			delete(ns.Clients, userID)
			ns.ClientsMu.Unlock()
			fmt.Printf("Client removed: %s\n", userID)
		}
	}
}

// handleBroadcasts broadcasts notifications to registered clients
func (ns *NotificationServer) handleBroadcasts() {
	for {
		select {
		case <-ns.Done:
			return

		case notification := <-ns.Broadcast:
			ns.broadcastNotification(notification)
		}
	}
}

// broadcastNotification sends a notification to all registered clients
func (ns *NotificationServer) broadcastNotification(notification Notification) {
	notifBytes, err := json.Marshal(notification)
	if err != nil {
		fmt.Printf("Failed to marshal notification: %v\n", err)
		return
	}

	ns.ClientsMu.RLock()
	clients := make([]*RegisteredClient, 0, len(ns.Clients))
	for _, client := range ns.Clients {
		clients = append(clients, client)
	}
	ns.ClientsMu.RUnlock()

	successCount := 0
	failedClients := []string{}

	for _, client := range clients {
		_, err := ns.Conn.WriteToUDP(notifBytes, client.Address)
		if err != nil {
			fmt.Printf("Failed to send notification to %s (%s): %v\n", client.UserID, client.Address.String(), err)
			failedClients = append(failedClients, client.UserID)
		} else {
			successCount++
		}
	}

	fmt.Printf("Notification broadcast: %s - Sent to %d clients, %d failed\n", 
		notification.MangaID, successCount, len(failedClients))

	// Remove failed clients (likely unreachable)
	if len(failedClients) > 0 {
		ns.ClientsMu.Lock()
		for _, userID := range failedClients {
			delete(ns.Clients, userID)
		}
		ns.ClientsMu.Unlock()
	}
}

// SendNotification queues a notification for broadcast
func (ns *NotificationServer) SendNotification(notification Notification) {
	select {
	case ns.Broadcast <- notification:
	default:
		fmt.Println("Notification queue full, dropping notification")
	}
}

// Stop gracefully shuts down the UDP server
func (ns *NotificationServer) Stop() {
	fmt.Println("Shutting down UDP notification server...")

	ns.ClientsMu.Lock()
	ns.Clients = make(map[string]*RegisteredClient)
	ns.ClientsMu.Unlock()

	if ns.Conn != nil {
		ns.Conn.Close()
	}

	close(ns.Done)
	fmt.Println("UDP notification server stopped")
}

// GetRegisteredClients returns the number of registered clients
func (ns *NotificationServer) GetRegisteredClients() int {
	ns.ClientsMu.RLock()
	defer ns.ClientsMu.RUnlock()
	return len(ns.Clients)
}

// GetClientList returns list of registered client IDs
func (ns *NotificationServer) GetClientList() []string {
	ns.ClientsMu.RLock()
	defer ns.ClientsMu.RUnlock()

	clients := make([]string, 0, len(ns.Clients))
	for userID := range ns.Clients {
		clients = append(clients, userID)
	}
	return clients
}

// IsClientRegistered checks if a client is registered
func (ns *NotificationServer) IsClientRegistered(userID string) bool {
	ns.ClientsMu.RLock()
	defer ns.ClientsMu.RUnlock()
	_, exists := ns.Clients[userID]
	return exists
}
