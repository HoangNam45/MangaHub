package dto

// AddToLibraryRequest represents the request to add manga to library
type AddToLibraryRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
}

// LibraryItem represents a manga in user's library
type LibraryItem struct {
	MangaID        string  `json:"manga_id"`
	Title          string  `json:"title"`
	Author         string  `json:"author"`
	Status         string  `json:"status"`
	Rating         float64 `json:"rating"`
	TotalChapters  int     `json:"total_chapters"`
	AddedAt        string  `json:"added_at"`
	CurrentChapter int     `json:"current_chapter"`
	Progress       int     `json:"progress"`
}

// GetLibraryResponse represents the response for getting user's library
type GetLibraryResponse struct {
	Library []LibraryItem `json:"library"`
	Count   int           `json:"count"`
}

// AddToLibraryResponse represents the response for adding to library
type AddToLibraryResponse struct {
	Message  string `json:"message"`
	MangaID  string `json:"manga_id"`
	AddedAt  string `json:"added_at"`
}

// UpdateProgressRequest represents the request to update reading progress
type UpdateProgressRequest struct {
	MangaID        string `json:"manga_id" binding:"required"`
	CurrentChapter int    `json:"current_chapter" binding:"required"`
}

// UpdateProgressResponse represents the response for updating progress
type UpdateProgressResponse struct {
	MangaID        string `json:"manga_id"`
	Title          string `json:"title"`
	CurrentChapter int    `json:"current_chapter"`
	TotalChapters  int    `json:"total_chapters"`
	Progress       int    `json:"progress"`
	UpdatedAt      string `json:"updated_at"`
	Message        string `json:"message"`
}
