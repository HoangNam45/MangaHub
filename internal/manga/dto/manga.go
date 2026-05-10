package dto

type SearchMangaRequest struct {
	Title  string `form:"title"`
	Author string `form:"author"`
	Genre  string `form:"genre"`
	Status string `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type MangaResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Genres      []string `json:"genres"`
	Status      string   `json:"status"`
	Chapters    int      `json:"chapters"`
	Rating      float64  `json:"rating"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type"` // "chapter_release" or "manga_update"
}

// NotificationResponse represents the response after sending a notification
type NotificationResponse struct {
	Message         string `json:"message"`
	MangaID         string `json:"manga_id"`
	ClientsNotified int    `json:"clients_notified"`
	Timestamp       string `json:"timestamp"`
}
