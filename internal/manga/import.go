package manga

import (
	"encoding/json"
	"fmt"
	"os"

	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

func ImportMangaFromJSON() {

	file, err := os.ReadFile("data/manga.json")
	if err != nil {
		panic(err)
	}

	var mangas []models.Manga
	err = json.Unmarshal(file, &mangas)
	if err != nil {
		panic(err)
	}


	for _, m := range mangas {
	
		genresJSON, _ := json.Marshal(m.Genres)

		query := `
		INSERT OR IGNORE INTO mangas 
		(id, title, description, author, genres, status, chapters, rating)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := database.DB.Exec(
			query,
			m.ID,
			m.Title,
			m.Description,
			m.Author,
			string(genresJSON),
			m.Status,
			m.Chapters,
			m.Rating,
		)

		if err != nil {
			fmt.Println("Insert error:", err)
		}
	}

	fmt.Println("Imported manga into DB successfully")
}