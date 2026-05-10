package main

import (
	"fmt"
	"net"

	pb "mangahub/internal/grpc/pb"
	"mangahub/internal/grpc/server"
	mangaService "mangahub/internal/manga/service"
	userService "mangahub/internal/user/service"
	"mangahub/pkg/database"

	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	database.InitDB()
	database.Migrate()

	// Create services
	ms := mangaService.NewMangaService()
	us := userService.NewUserLibraryService()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create and register the MangaService
	mangaServer := server.NewMangaServer(ms, us)
	pb.RegisterMangaServiceServer(grpcServer, mangaServer)

	// Listen on port 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("Failed to listen on port 50051: %v\n", err)
		return
	}

	fmt.Println("gRPC Server is running on port 50051")
	fmt.Println("Services registered:")
	fmt.Println("  - MangaService with GetManga, SearchManga, UpdateProgress")

	// Start the gRPC server
	if err := grpcServer.Serve(listener); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
	}
}
