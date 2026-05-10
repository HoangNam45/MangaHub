package service

import (
	"time"

	"mangahub/internal/user/dto"
	"mangahub/internal/user/repository"
)

type UserLibraryService struct {
	repo *repository.UserLibraryRepository
}

// NewUserLibraryService creates a new instance of UserLibraryService
func NewUserLibraryService() *UserLibraryService {
	return &UserLibraryService{
		repo: &repository.UserLibraryRepository{},
	}
}

// AddToLibrary adds a manga to the user's library
func (uls *UserLibraryService) AddToLibrary(userID, mangaID string) (*dto.AddToLibraryResponse, error) {
	_, err := uls.repo.AddToLibrary(userID, mangaID)
	if err != nil {
		return nil, err
	}

	return &dto.AddToLibraryResponse{
		Message:  "Manga added to library successfully",
		MangaID:  mangaID,
		AddedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// GetUserLibrary retrieves the user's library with all manga and their reading progress
func (uls *UserLibraryService) GetUserLibrary(userID string) (*dto.GetLibraryResponse, error) {
	library, err := uls.repo.GetUserLibrary(userID)
	if err != nil {
		return nil, err
	}

	items := convertToLibraryItems(library)

	return &dto.GetLibraryResponse{
		Library: items,
		Count:   len(items),
	}, nil
}

// RemoveFromLibrary removes a manga from the user's library
func (uls *UserLibraryService) RemoveFromLibrary(userID, mangaID string) error {
	return uls.repo.RemoveFromLibrary(userID, mangaID)
}

// UpdateReadingProgress updates the user's reading progress for a manga
// Progress is automatically calculated based on currentChapter and total chapters
func (uls *UserLibraryService) UpdateReadingProgress(userID, mangaID string, currentChapter int) (*dto.UpdateProgressResponse, error) {
	result, err := uls.repo.UpdateReadingProgress(userID, mangaID, currentChapter)
	if err != nil {
		return nil, err
	}

	return &dto.UpdateProgressResponse{
		MangaID:        mangaID,
		Title:          result["title"].(string),
		CurrentChapter: result["current_chapter"].(int),
		TotalChapters:  result["total_chapters"].(int),
		Progress:       result["progress"].(int),
		UpdatedAt:      time.Now().Format(time.RFC3339),
		Message:        "Reading progress updated successfully",
	}, nil
}

// Helper function to convert library data to DTOs
func convertToLibraryItems(library []map[string]interface{}) []dto.LibraryItem {
	if library == nil {
		return []dto.LibraryItem{}
	}

	items := make([]dto.LibraryItem, 0, len(library))

	for _, item := range library {
		libraryItem := dto.LibraryItem{
			MangaID: item["manga_id"].(string),
			Title:   item["title"].(string),
			Author:  item["author"].(string),
			Status:  item["status"].(string),
		}

		// Handle rating
		if rating, ok := item["rating"].(float64); ok {
			libraryItem.Rating = rating
		}

		// Handle total_chapters
		if totalChapters, ok := item["total_chapters"].(int); ok {
			libraryItem.TotalChapters = totalChapters
		}

		// Handle added_at
		if addedAt, ok := item["added_at"].(string); ok {
			libraryItem.AddedAt = addedAt
		}

		// Handle progress
		if progress, ok := item["progress"].(int); ok {
			libraryItem.Progress = progress
		}

		// Handle current_chapter
		if currentChapter, ok := item["current_chapter"].(int); ok {
			libraryItem.CurrentChapter = currentChapter
		}

		items = append(items, libraryItem)
	}

	return items
}
