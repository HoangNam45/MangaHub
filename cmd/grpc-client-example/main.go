package main

import (
	"fmt"
	"log"

	"mangahub/internal/grpc/client"
)

func main() {
	// Create a new gRPC client
	grpcClient, err := client.NewMangaClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v\n", err)
	}
	defer grpcClient.Close()

	fmt.Println("=== gRPC Client Example ===\n")

	// Example 1: GetManga
	fmt.Println("1. GetManga Example:")
	mangaResp, err := grpcClient.GetManga("manga-001")
	if err != nil {
		log.Printf("Error getting manga: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", mangaResp.Success)
		fmt.Printf("Message: %s\n", mangaResp.Message)
		if mangaResp.Manga != nil {
			fmt.Printf("Manga: %s by %s (%s)\n", mangaResp.Manga.Title, mangaResp.Manga.Author, mangaResp.Manga.Status)
			fmt.Printf("Rating: %.1f/10\n", mangaResp.Manga.Rating)
		}
	}

	fmt.Println("\n2. SearchManga Example:")
	searchResp, err := grpcClient.SearchManga("", "", "", "", 1, 10)
	if err != nil {
		log.Printf("Error searching manga: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", searchResp.Success)
		fmt.Printf("Total Results: %d\n", searchResp.TotalCount)
		fmt.Printf("Results on Page %d: %d\n", searchResp.Page, len(searchResp.Results))
		for i, manga := range searchResp.Results {
			fmt.Printf("  %d. %s - %d chapters\n", i+1, manga.Title, manga.Chapters)
		}
	}

	fmt.Println("\n3. UpdateProgress Example:")
	progressResp, err := grpcClient.UpdateProgress("user-001", "manga-001", 5)
	if err != nil {
		log.Printf("Error updating progress: %v\n", err)
	} else {
		fmt.Printf("Success: %v\n", progressResp.Success)
		fmt.Printf("Message: %s\n", progressResp.Message)
		fmt.Printf("Reading: %s\n", progressResp.Title)
		fmt.Printf("Progress: Chapter %d/%d (%d%%)\n", progressResp.CurrentChapter, progressResp.TotalChapters, progressResp.Progress)
		fmt.Printf("Updated At: %s\n", progressResp.UpdatedAt)
	}
}
