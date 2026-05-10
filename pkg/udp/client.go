package udp

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// ClientExample represents a UDP notification client
type ClientExample struct {
	UserID      string
	Preferences map[string]interface{}
	Conn        *net.UDPConn
	ServerAddr  *net.UDPAddr
}

// ConnectToNotificationServer connects to the UDP notification server
func ConnectToNotificationServer(serverHost string, serverPort string, userID string, preferences map[string]interface{}) (*ClientExample, error) {
	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", serverHost+":"+serverPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve server address: %v", err)
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	client := &ClientExample{
		UserID:      userID,
		Preferences: preferences,
		Conn:        conn,
		ServerAddr:  serverAddr,
	}

	// Send registration message
	err = client.Register()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return client, nil
}

// Register sends a registration message to the server
func (ce *ClientExample) Register() error {
	registration := ClientRegistration{
		Type:        "register",
		UserID:      ce.UserID,
		Preferences: ce.Preferences,
	}

	regBytes, err := json.Marshal(registration)
	if err != nil {
		return err
	}

	// Send registration
	_, err = ce.Conn.Write(regBytes)
	if err != nil {
		return err
	}

	// Wait for registration confirmation
	buffer := make([]byte, 4096)
	err = ce.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return err
	}

	n, err := ce.Conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to receive registration confirmation: %v", err)
	}

	var response RegistrationResponse
	err = json.Unmarshal(buffer[:n], &response)
	if err != nil {
		return err
	}

	if response.Status != "success" {
		return fmt.Errorf("registration failed: %s", response.Message)
	}

	fmt.Printf("Registration confirmed: %s\n", response.Message)

	// Clear deadline for normal reading
	err = ce.Conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}

	return nil
}

// ListenForNotifications listens for incoming notifications
func (ce *ClientExample) ListenForNotifications(done chan bool) {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-done:
			fmt.Println("Stopped listening for notifications")
			return
		default:
			n, err := ce.Conn.Read(buffer)
			if err != nil {
				if err.(net.Error).Timeout() {
					continue
				}
				fmt.Printf("Connection error: %v\n", err)
				close(done)
				return
			}

			var notification Notification
			err = json.Unmarshal(buffer[:n], &notification)
			if err != nil {
				fmt.Printf("Failed to parse notification: %v\n", err)
				continue
			}

			// Display notification
			fmt.Printf("\nNOTIFICATION - %s\n", notification.Type)
			fmt.Printf("   Manga: %s\n", notification.Title)
			fmt.Printf("   Message: %s\n", notification.Message)
			fmt.Printf("   MangaID: %s\n", notification.MangaID)
			fmt.Printf("   Timestamp: %s\n\n", time.Unix(notification.Timestamp, 0).Format(time.RFC3339))
		}
	}
}

// Close closes the connection
func (ce *ClientExample) Close() error {
	return ce.Conn.Close()
}
