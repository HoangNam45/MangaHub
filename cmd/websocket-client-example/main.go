package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

// ChatMessage represents a message sent through the chat
type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "message", "join", "leave", "system", "error"
}

func main() {
	// Parse command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: websocket-client-example <server_url> <user_id> <username>")
		fmt.Println("Example: websocket-client-example ws://localhost:8080/ws/chat user123 John")
		os.Exit(1)
	}

	wsURL := os.Args[1]
	userID := os.Args[2]
	username := os.Args[3]

	// Dial the WebSocket server
	fmt.Printf("Connecting to %s as %s...\n", wsURL, username)
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Send join message
	joinMsg := map[string]string{
		"user_id":  userID,
		"username": username,
	}

	err = conn.WriteJSON(joinMsg)
	if err != nil {
		log.Fatalf("Failed to send join message: %v", err)
	}

	fmt.Printf("Connected as %s (%s)\n", username, userID)
	fmt.Println("Type your messages below. Type 'exit' to disconnect.")
	fmt.Println("---")

	// Start reading messages from server in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var msg ChatMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Display the message
			displayMessage(msg)
		}
	}()

	// Read messages from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		if text == "exit" {
			fmt.Println("Disconnecting...")
			conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if text == "" {
			continue
		}

		// Send message to server
		msg := ChatMessage{
			Message: text,
		}

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}

	<-done
}

func displayMessage(msg ChatMessage) {
	switch msg.Type {
	case "join":
		fmt.Printf("\n[SYSTEM] %s joined the chat\n> ", msg.Username)
	case "leave":
		fmt.Printf("\n[SYSTEM] %s left the chat\n> ", msg.Username)
	case "error":
		fmt.Printf("\n[ERROR] %s\n> ", msg.Message)
	case "system":
		fmt.Printf("\n[SYSTEM] %s\n> ", msg.Message)
	case "message":
		fmt.Printf("\n[%s]: %s\n> ", msg.Username, msg.Message)
	default:
		fmt.Printf("\n[%s]: %s\n> ", msg.Username, msg.Message)
	}
}
