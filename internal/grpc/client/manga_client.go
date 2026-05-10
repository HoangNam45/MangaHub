package client

import (
	"context"
	"fmt"
	"time"

	pb "mangahub/internal/grpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


type MangaClient struct {
	client pb.MangaServiceClient
	conn   *grpc.ClientConn
}

func NewMangaClient(addr string) (*MangaClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &MangaClient{
		client: pb.NewMangaServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetManga calls the GetManga RPC method
func (mc *MangaClient) GetManga(mangaID string) (*pb.MangaResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.GetMangaRequest{
		MangaId: mangaID,
	}

	resp, err := mc.client.GetManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetManga RPC failed: %w", err)
	}

	return resp, nil
}

// SearchManga calls the SearchManga RPC method
func (mc *MangaClient) SearchManga(title, author, genre, status string, page, pageSize int32) (*pb.SearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchRequest{
		Title:    title,
		Author:   author,
		Genre:    genre,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := mc.client.SearchManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SearchManga RPC failed: %w", err)
	}

	return resp, nil
}

// UpdateProgress calls the UpdateProgress RPC method
func (mc *MangaClient) UpdateProgress(userID, mangaID string, currentChapter int32) (*pb.ProgressResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.ProgressRequest{
		UserId:           userID,
		MangaId:          mangaID,
		CurrentChapter:   currentChapter,
	}

	resp, err := mc.client.UpdateProgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateProgress RPC failed: %w", err)
	}

	return resp, nil
}

// Close closes the client connection
func (mc *MangaClient) Close() error {
	return mc.conn.Close()
}
