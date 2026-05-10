package repository

import (
	"encoding/json"
	"errors"
	"strings"

	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

type MangaRepository struct{}

// GetAllManga retrieves all manga from database
func (mr *MangaRepository) GetAllManga() ([]*models.Manga, error) {
	query := `SELECT id, title, description, author, genres, status, chapters, rating FROM mangas`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []*models.Manga
	for rows.Next() {
		manga := &models.Manga{}
		var genresJSON string
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Description, &manga.Author, &genresJSON, &manga.Status, &manga.Chapters, &manga.Rating)
		if err != nil {
			return nil, err
		}

		// Parse genres from JSON
		err = json.Unmarshal([]byte(genresJSON), &manga.Genres)
		if err != nil {
			return nil, err
		}

		mangas = append(mangas, manga)
	}

	return mangas, rows.Err()
}

// GetMangaByID retrieves a manga by ID using direct query
func (mr *MangaRepository) GetMangaByID(id string) (*models.Manga, error) {
	query := `SELECT id, title, description, author, genres, status, chapters, rating FROM mangas WHERE id = ?`
	manga := &models.Manga{}
	var genresJSON string

	err := database.DB.QueryRow(query, id).Scan(&manga.ID, &manga.Title, &manga.Description, &manga.Author, &genresJSON, &manga.Status, &manga.Chapters, &manga.Rating)
	if err != nil {
		return nil, errors.New("manga not found")
	}

	// Parse genres from JSON
	err = json.Unmarshal([]byte(genresJSON), &manga.Genres)
	if err != nil {
		return nil, err
	}

	return manga, nil
}

// SearchManga searches for manga based on filters using SQL WHERE clauses
func (mr *MangaRepository) SearchManga(title, author, genre, status string) ([]*models.Manga, error) {
	query := `SELECT id, title, description, author, genres, status, chapters, rating FROM mangas WHERE 1=1`
	var args []interface{}

	// Build dynamic WHERE clauses
	if title != "" {
		query += ` AND LOWER(title) LIKE ?`
		args = append(args, "%"+strings.ToLower(title)+"%")
	}

	if author != "" {
		query += ` AND LOWER(author) LIKE ?`
		args = append(args, "%"+strings.ToLower(author)+"%")
	}

	if status != "" {
		query += ` AND LOWER(status) = ?`
		args = append(args, strings.ToLower(status))
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.Manga
	for rows.Next() {
		manga := &models.Manga{}
		var genresJSON string
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Description, &manga.Author, &genresJSON, &manga.Status, &manga.Chapters, &manga.Rating)
		if err != nil {
			return nil, err
		}

		// Parse genres from JSON
		err = json.Unmarshal([]byte(genresJSON), &manga.Genres)
		if err != nil {
			return nil, err
		}

		// Filter by genre in memory only if specified (since genres is JSON array)
		if genre != "" {
			hasGenre := false
			for _, g := range manga.Genres {
				if strings.EqualFold(g, genre) {
					hasGenre = true
					break
				}
			}
			if !hasGenre {
				continue
			}
		}

		results = append(results, manga)
	}

	return results, rows.Err()
}
