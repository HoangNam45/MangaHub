package repository

import (
	"database/sql"
	"errors"

	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

type UserLibraryRepository struct{}

// AddToLibrary adds a manga to user's library
func (ulr *UserLibraryRepository) AddToLibrary(userID, mangaID string) (string, error) {
	if userID == "" || mangaID == "" {
		return "", errors.New("user_id and manga_id are required")
	}

	query := `
	INSERT INTO user_library (id, user_id, manga_id, added_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	// Generate a simple ID
	id := userID + "_" + mangaID

	_, err := database.DB.Exec(query, id, userID, mangaID)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: user_library.user_id, user_library.manga_id" {
			return "", errors.New("manga already in library")
		}
		return "", err
	}

	return id, nil
}

// GetUserLibrary retrieves all manga in user's library with reading progress
func (ulr *UserLibraryRepository) GetUserLibrary(userID string) ([]map[string]interface{}, error) {
	query := `
	SELECT 
		m.id,
		m.title,
		m.author,
		m.status,
		m.rating,
		m.chapters as total_chapters,
		ul.added_at,
		COALESCE(rp.current_chapter, 0) as current_chapter,
		COALESCE(rp.progress, 0) as progress
	FROM user_library ul
	JOIN mangas m ON ul.manga_id = m.id
	LEFT JOIN reading_progress rp ON ul.user_id = rp.user_id AND ul.manga_id = rp.manga_id
	WHERE ul.user_id = ?
	ORDER BY ul.added_at DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var library []map[string]interface{}

	for rows.Next() {
		var id, title, author, status string
		var rating float64
		var totalChapters, currentChapter, progress int
		var addedAt string

		err := rows.Scan(&id, &title, &author, &status, &rating, &totalChapters, &addedAt, &currentChapter, &progress)
		if err != nil {
			return nil, err
		}

		item := map[string]interface{}{
			"manga_id":         id,
			"title":            title,
			"author":           author,
			"status":           status,
			"rating":           rating,
			"total_chapters":   totalChapters,
			"added_at":         addedAt,
			"current_chapter": currentChapter,
			"progress":         progress,
		}

		library = append(library, item)
	}

	return library, nil
}

// RemoveFromLibrary removes a manga from user's library
func (ulr *UserLibraryRepository) RemoveFromLibrary(userID, mangaID string) error {
	query := `DELETE FROM user_library WHERE user_id = ? AND manga_id = ?`

	result, err := database.DB.Exec(query, userID, mangaID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("manga not found in library")
	}

	return nil
}

// UpdateReadingProgress updates the user's reading progress for a manga
// It calculates progress automatically based on currentChapter and manga's total chapters
func (ulr *UserLibraryRepository) UpdateReadingProgress(userID, mangaID string, currentChapter int) (map[string]interface{}, error) {
	// First check if user has the manga in library and get manga details
	checkQuery := `SELECT ul.id, m.chapters, m.title FROM user_library ul 
	JOIN mangas m ON ul.manga_id = m.id 
	WHERE ul.user_id = ? AND ul.manga_id = ?`
	var libraryID, title string
	var totalChapters int
	err := database.DB.QueryRow(checkQuery, userID, mangaID).Scan(&libraryID, &totalChapters, &title)
	if err == sql.ErrNoRows {
		return nil, errors.New("manga not in user's library")
	}
	if err != nil {
		return nil, err
	}

	// Calculate progress automatically
	var calculatedProgress int
	if totalChapters > 0 {
		calculatedProgress = (currentChapter * 100) / totalChapters
	}

	// Upsert reading progress
	query := `
	INSERT INTO reading_progress (id, user_id, manga_id, current_chapter, progress, updated_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id, manga_id) DO UPDATE SET
		current_chapter = excluded.current_chapter,
		progress = excluded.progress,
		updated_at = CURRENT_TIMESTAMP
	`

	// Generate ID for new records
	id := userID + "_progress_" + mangaID

	_, err = database.DB.Exec(query, id, userID, mangaID, currentChapter, calculatedProgress)
	if err != nil {
		return nil, err
	}

	// Return the result map
	result := map[string]interface{}{
		"manga_id":       mangaID,
		"title":          title,
		"current_chapter": currentChapter,
		"total_chapters":  totalChapters,
		"progress":        calculatedProgress,
	}

	return result, nil
}

// GetReadingProgress retrieves reading progress for a specific manga
func (ulr *UserLibraryRepository) GetReadingProgress(userID, mangaID string) (*models.ReadingProgress, error) {
	query := `SELECT id, user_id, manga_id, current_chapter, progress, updated_at FROM reading_progress WHERE user_id = ? AND manga_id = ?`

	progress := &models.ReadingProgress{}
	err := database.DB.QueryRow(query, userID, mangaID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.MangaID,
		&progress.CurrentChapter,
		&progress.Progress,
		&progress.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("reading progress not found")
	}
	if err != nil {
		return nil, err
	}

	return progress, nil
}
