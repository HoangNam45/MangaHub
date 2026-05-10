package server

import (
	"context"
	"fmt"

	pb "mangahub/internal/grpc/pb"
	"mangahub/internal/manga/service"
	userservice "mangahub/internal/user/service"
	"mangahub/pkg/tcp"
)

// MangaServer implements the gRPC MangaService
type MangaServer struct {
	pb.UnimplementedMangaServiceServer
	mangaService *service.MangaService
	userService  *userservice.UserLibraryService
}

func NewMangaServer(mangaService *service.MangaService, userService *userservice.UserLibraryService) *MangaServer {
	return &MangaServer{
		mangaService: mangaService,
		userService:  userService,
	}
}

func (s *MangaServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.MangaResponse, error) {
	// Validate request
	if req.MangaId == "" {
		return &pb.MangaResponse{
			Success:   false,
			Message:   "Manga ID is required",
			ErrorCode: 400,
		}, nil
	}

	// Retrieve manga from service
	manga, err := s.mangaService.GetMangaByID(req.MangaId)
	if err != nil {
		return &pb.MangaResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to retrieve manga: %v", err),
			ErrorCode: 500,
		}, nil
	}

	if manga == nil {
		return &pb.MangaResponse{
			Success:   false,
			Message:   "Manga not found",
			ErrorCode: 404,
		}, nil
	}

	// Convert to protobuf message
	pbManga := &pb.Manga{
		Id:          manga.ID,
		Title:       manga.Title,
		Description: manga.Description,
		Author:      manga.Author,
		Genres:      manga.Genres,
		Status:      manga.Status,
		Chapters:    int32(manga.Chapters),
		Rating:      manga.Rating,
	}

	return &pb.MangaResponse{
		Success: true,
		Message: "Manga retrieved successfully",
		Manga:   pbManga,
	}, nil
}

func (s *MangaServer) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	// Set default page values
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // default page size
	}

	// Search manga using the service
	mangaList, err := s.mangaService.SearchManga(
		req.Title,
		req.Author,
		req.Genre,
		req.Status,
	)
	if err != nil {
		return &pb.SearchResponse{
			Success:   false,
			Message:   fmt.Sprintf("Search failed: %v", err),
			ErrorCode: 500,
		}, nil
	}

	// Apply pagination
	totalCount := int32(len(mangaList))
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= int32(len(mangaList)) {
		// Return empty results for out-of-range pages
		return &pb.SearchResponse{
			Success:    true,
			Message:    "No results found",
			Results:    []*pb.Manga{},
			TotalCount: totalCount,
			Page:       page,
		}, nil
	}

	if endIdx > int32(len(mangaList)) {
		endIdx = int32(len(mangaList))
	}

	// Convert to protobuf messages
	results := make([]*pb.Manga, 0, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		if mangaList[i] != nil {
			pbManga := &pb.Manga{
				Id:          mangaList[i].ID,
				Title:       mangaList[i].Title,
				Description: mangaList[i].Description,
				Author:      mangaList[i].Author,
				Genres:      mangaList[i].Genres,
				Status:      mangaList[i].Status,
				Chapters:    int32(mangaList[i].Chapters),
				Rating:      mangaList[i].Rating,
			}
			results = append(results, pbManga)
		}
	}

	return &pb.SearchResponse{
		Success:    true,
		Message:    "Search completed successfully",
		Results:    results,
		TotalCount: totalCount,
		Page:       page,
	}, nil
}

func (s *MangaServer) UpdateProgress(ctx context.Context, req *pb.ProgressRequest) (*pb.ProgressResponse, error) {
	// Validate request
	if req.UserId == "" {
		return &pb.ProgressResponse{
			Success:   false,
			Message:   "User ID is required",
			ErrorCode: 400,
		}, nil
	}

	if req.MangaId == "" {
		return &pb.ProgressResponse{
			Success:   false,
			Message:   "Manga ID is required",
			ErrorCode: 400,
		}, nil
	}

	if req.CurrentChapter < 0 {
		return &pb.ProgressResponse{
			Success:   false,
			Message:   "Current chapter cannot be negative",
			ErrorCode: 400,
		}, nil
	}

	// Update progress using the service
	result, err := s.userService.UpdateReadingProgress(req.UserId, req.MangaId, int(req.CurrentChapter))
	if err != nil {
		return &pb.ProgressResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to update progress: %v", err),
			ErrorCode: 500,
		}, nil
	}

	// Trigger TCP broadcast for real-time sync (mentioned in UC-016)
	go func() {
		tcpServer := tcp.GetGlobalServer()
		if tcpServer != nil {
			// Construct progress update message
			progressUpdate := tcp.ProgressUpdate{
				UserID:    req.UserId,
				MangaID:   req.MangaId,
				Chapter:   int(req.CurrentChapter),
				Timestamp: 0, // Will be set by TCP server if needed
			}
			// Send to broadcast channel
			select {
			case tcpServer.Broadcast <- progressUpdate:
			default:
				// Channel full, skip broadcast
			}
		}
	}()

	return &pb.ProgressResponse{
		Success:        true,
		Message:        "Progress updated successfully",
		MangaId:        req.MangaId,
		Title:          result.Title,
		CurrentChapter: int32(result.CurrentChapter),
		TotalChapters:  int32(result.TotalChapters),
		Progress:       int32(result.Progress),
		UpdatedAt:      result.UpdatedAt,
	}, nil
}
