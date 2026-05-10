package tcp

import (
	"encoding/json"
	"net"
	"time"
)

// ClientExample demonstrates how to connect to the TCP server
// This is useful for testing and client implementations
type ClientExample struct {
	Conn net.Conn
}

// ConnectToServer connects to the TCP server
func ConnectToServer(serverAddr string, userID, username, token string) (*ClientExample, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	// Send authentication message
	authMsg := AuthMessage{
		Type:     "auth",
		UserID:   userID,
		Username: username,
		Token:    token,
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(authMsg)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Wait for authentication response
	decoder := json.NewDecoder(conn)
	var response ServerMessage
	err = decoder.Decode(&response)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if response.Status != "success" {
		conn.Close()
		return nil, err
	}

	return &ClientExample{Conn: conn}, nil
}

// ListenForUpdates listens for progress updates from the server
func (ce *ClientExample) ListenForUpdates(done chan bool) {
	decoder := json.NewDecoder(ce.Conn)

	for {
		select {
		case <-done:
			return
		default:
			var msg ServerMessage
			err := decoder.Decode(&msg)
			if err != nil {
				close(done)
				return
			}

			if msg.Type == "progress_update" {
				var update ProgressUpdate
				err := json.Unmarshal(msg.Payload, &update)
				if err != nil {
					continue
				}
				// Handle the update (e.g., update UI, log, etc.)
			}
		}
	}
}

// SendPing sends a ping message to keep the connection alive
func (ce *ClientExample) SendPing() error {
	msg := map[string]string{
		"type": "ping",
	}
	encoder := json.NewEncoder(ce.Conn)
	return encoder.Encode(msg)
}

// KeepAlive sends periodic pings to keep the connection alive
func (ce *ClientExample) KeepAlive(interval time.Duration, done chan bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ce.SendPing()
		}
	}
}

// Close closes the connection
func (ce *ClientExample) Close() error {
	return ce.Conn.Close()
}
