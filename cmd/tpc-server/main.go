package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mangahub/pkg/tcp"
)

func main() {
	// Create and start TCP server
	tcpServer := tcp.NewProgressSyncServer("9090")

	err := tcpServer.Start()
	if err != nil {
		fmt.Printf("Failed to start TCP server: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutdown signal received")

	tcpServer.Stop()
	os.Exit(0)
}
