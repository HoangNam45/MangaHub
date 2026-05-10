package main

import (
	"encoding/json"
	"fmt"
	"time"

	"mangahub/pkg/tcp"
)

// This is an example client that connects to the TCP server
// Run the API server first (cmd/api-server), then this example

func main() {
	// Example: Connect to the TCP server
	serverAddr := "localhost:9090"
	userID := "76ad735e-4546-47ff-814f-383fd57d14b5"
	username := "nam123"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNzZhZDczNWUtNDU0Ni00N2ZmLTgxNGYtMzgzZmQ1N2QxNGI1IiwidXNlcm5hbWUiOiJuYW0xMjMiLCJleHAiOjE3NzU3Mjc2MTksImlhdCI6MTc3NTY0MTIxOX0.Ef6n0_JPNa95ABDcmDi7LaOq2Eo5iXT6hUcpROTyyEY"

	fmt.Printf("Connecting to TCP server at %s...\n", serverAddr)

	client, err := tcp.ConnectToServer(serverAddr, userID, username, token)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	defer client.Close()
	fmt.Println("Connected successfully!")

	// Create channels for coordination
	done := make(chan bool)
	updateChan := make(chan tcp.ProgressUpdate, 10)

	// Listen for updates in background
	go listenForUpdates(client, updateChan, done)

	// Keep connection alive with periodic pings
	go client.KeepAlive(30*time.Second, done)

	// Simulate receiving and processing updates
	fmt.Println("Listening for progress updates...")
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			fmt.Println("Timeout reached, closing connection")
			close(done)
			return
		case update := <-updateChan:
			fmt.Printf("Received update: User %s reading manga %s at chapter %d (timestamp: %d)\n",
				update.UserID, update.MangaID, update.Chapter, update.Timestamp)
		}
	}
}

func listenForUpdates(client *tcp.ClientExample, updateChan chan tcp.ProgressUpdate, done chan bool) {
	decoder := json.NewDecoder(client.Conn)

	for {
		select {
		case <-done:
			return
		default:
			var msg tcp.ServerMessage
			err := decoder.Decode(&msg)
			if err != nil {
				fmt.Printf("Connection error: %v\n", err)
				close(done)
				return
			}

			switch msg.Type {
			case "progress_update":
				var update tcp.ProgressUpdate
				err := json.Unmarshal(msg.Payload, &update)
				if err != nil {
					fmt.Printf("Failed to parse update: %v\n", err)
					continue
				}
				updateChan <- update

			case "pong":
				fmt.Println("Pong received - connection alive")

			case "error":
				fmt.Printf("Server error: %s\n", msg.Message)

			default:
				fmt.Printf("Received: %s - %s\n", msg.Type, msg.Message)
			}
		}
	}
}
