package service

import (
	"mangahub/internal/manga/repository"
	"mangahub/pkg/models"
)

type MangaService struct {
	repo *repository.MangaRepository
}

// NewMangaService creates a new instance of MangaService
func NewMangaService() *MangaService {
	return &MangaService{
		repo: &repository.MangaRepository{},
	}
}

// GetMangaByID retrieves a manga by ID
func (ms *MangaService) GetMangaByID(id string) (*models.Manga, error) {
	return ms.repo.GetMangaByID(id)
}

// SearchManga searches for manga based on filters
func (ms *MangaService) SearchManga(title, author, genre, status string) ([]*models.Manga, error) {
	return ms.repo.SearchManga(title, author, genre, status)
}
