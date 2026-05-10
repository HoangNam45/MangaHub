package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub/pkg/udp"
)

func main() {
	// Create and start UDP notification server
	udpServer := udp.NewNotificationServer("9091")

	err := udpServer.Start()
	if err != nil {
		fmt.Printf("Failed to start UDP server: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Display server status every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				return
			case <-ticker.C:
				fmt.Printf("Active clients: %d\n", udpServer.GetRegisteredClients())
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutdown signal received")

	udpServer.Stop()
	os.Exit(0)
}
