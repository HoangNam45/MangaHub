package models

type Manga struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Genres      []string `json:"genres"`
	Status      string   `json:"status"`
	Chapters    int      `json:"chapters"`
	Rating      float64  `json:"rating"`
}