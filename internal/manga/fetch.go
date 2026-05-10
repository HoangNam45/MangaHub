package manga

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"mangahub/pkg/models"
)

type MangaDexResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title         map[string]string   `json:"title"`
			AltTitles     []map[string]string `json:"altTitles"`
			Description   map[string]string   `json:"description"`
			Status        string              `json:"status"`
			ContentRating string              `json:"contentRating"`
			Tags          []struct {
				Attributes struct {
					Name map[string]string `json:"name"`
				} `json:"attributes"`
			} `json:"tags"`
		} `json:"attributes"`
		Relationships []struct {
			Type       string `json:"type"`
			Attributes struct {
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"relationships"`
	} `json:"data"`
}

func FetchAndAppendManga() {

	file, err := os.ReadFile("data/manga.json")
	if err != nil {
		panic(err)
	}

	var mangas []models.Manga
	json.Unmarshal(file, &mangas)

	idCounter := 101
	limit := 50
	offset := 0

	for idCounter <= 200 {

		url := fmt.Sprintf("https://api.mangadex.org/manga?limit=%d&offset=%d", limit, offset)

		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		var apiRes MangaDexResponse
		json.NewDecoder(resp.Body).Decode(&apiRes)
		resp.Body.Close()


		if len(apiRes.Data) == 0 {
			break
		}

		for _, item := range apiRes.Data {

	
			if item.Attributes.ContentRating == "erotica" || item.Attributes.ContentRating == "pornographic" {
				continue
			}

		
			title := item.Attributes.Title["en"]

			if title == "" {
				for _, t := range item.Attributes.Title {
					title = t
					break
				}
			}

			if title == "" {
				for _, alt := range item.Attributes.AltTitles {
					if en, ok := alt["en"]; ok {
						title = en
						break
					}
				}
			}

			if title == "" {
				continue
			}

	
			desc := item.Attributes.Description["en"]
			if desc == "" {
				desc = "No description available"
			}

		
			author := "Unknown"
			for _, rel := range item.Relationships {
				if rel.Type == "author" {
					author = rel.Attributes.Name
					break
				}
			}

		
			var genres []string
			for _, tag := range item.Attributes.Tags {
				if name, ok := tag.Attributes.Name["en"]; ok {
					genres = append(genres, name)
				}
			}
			if len(genres) == 0 {
				genres = []string{"Unknown"}
			}

	
			manga := models.Manga{
				ID:          fmt.Sprintf("%d", idCounter),
				Title:       title,
				Description: desc,
				Author:      author,
				Genres:      genres,
				Status:      item.Attributes.Status,
				Chapters:    0,
				Rating:      0.0,
			}

			mangas = append(mangas, manga)
			idCounter++

			if idCounter > 200 {
				break
			}
		}

		offset += limit
	}


	newFile, err := os.Create("data/manga.json")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()

	encoder := json.NewEncoder(newFile)
	encoder.SetIndent("", "  ")
	encoder.Encode(mangas)

	fmt.Println("Successfully added MangaDex data (101–200)")
}