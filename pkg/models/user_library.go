package models

// UserLibrary represents a manga in a user's library
type UserLibrary struct {
	ID      string
	UserID  string
	MangaID string
	AddedAt string
}

// ReadingProgress represents the user's reading progress for a manga
type ReadingProgress struct {
	ID             string
	UserID         string
	MangaID        string
	CurrentChapter int
	Progress       int // percentage: 0-100
	UpdatedAt      string
}
