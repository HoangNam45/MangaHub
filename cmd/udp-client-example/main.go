package main

import (
	"fmt"
	"time"

	"mangahub/pkg/udp"
)


func main() {
	// Connect to the UDP notification server
	serverHost := "localhost"
	serverPort := "9091"
	userID := "user_124"

	// Define preferences (optional)
	preferences := map[string]interface{}{
		"genres": []string{"action", "adventure", "fantasy"},
	}

	fmt.Printf("Connecting to UDP notification server at %s:%s...\n", serverHost, serverPort)

	client, err := udp.ConnectToNotificationServer(serverHost, serverPort, userID, preferences)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	defer client.Close()

	// Create a done channel for graceful shutdown
	done := make(chan bool)

	// Listen for notifications in background
	go client.ListenForNotifications(done)

	// Simulate timeout after 5 minutes
	timeout := time.After(5 * time.Minute)

	fmt.Println("Listening for chapter release notifications...")
	fmt.Println("Waiting for notifications (Ctrl+C to exit)...")

	select {
	case <-timeout:
		fmt.Println("\nTimeout reached, disconnecting")
		close(done)
	}
}
