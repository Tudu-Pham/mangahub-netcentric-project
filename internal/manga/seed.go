package manga

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"
)

type MangaJSON struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Genres        []string `json:"genres"`
	Status        string   `json:"status"`
	TotalChapters int      `json:"total_chapters"`
	Description   string   `json:"description"`
	CoverURL      string   `json:"cover_url"`
}

func SeedMangaFromJSON(db *sql.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var mangas []MangaJSON
	if err := json.Unmarshal(data, &mangas); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO manga 
		(id, title, author, genres, status, total_chapters, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, m := range mangas {
		genresText := strings.Join(m.Genres, ",")

		_, err := stmt.Exec(
			m.ID,
			m.Title,
			m.Author,
			genresText,
			m.Status,
			m.TotalChapters,
			m.Description,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
