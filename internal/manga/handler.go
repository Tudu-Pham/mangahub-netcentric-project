package manga

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/manga", h.GetAllManga)
	r.GET("/manga/:id", h.GetMangaByID)
}

func (h *Handler) GetAllManga(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, title, author FROM manga")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	mangas := []gin.H{}

	for rows.Next() {
		var id, title, author string
		if err := rows.Scan(&id, &title, &author); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		mangas = append(mangas, gin.H{
			"id":     id,
			"title":  title,
			"author": author,
		})
	}

	c.JSON(http.StatusOK, mangas)
}

func (h *Handler) GetMangaByID(c *gin.Context) {
	id := c.Param("id")

	var manga struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		Author        string `json:"author"`
		Genres        string `json:"genres"`
		Status        string `json:"status"`
		TotalChapters int    `json:"total_chapters"`
		Description   string `json:"description"`
	}

	err := h.DB.QueryRow(`
		SELECT id, title, author, genres, status, total_chapters, description
		FROM manga
		WHERE id = ?
	`, id).Scan(
		&manga.ID,
		&manga.Title,
		&manga.Author,
		&manga.Genres,
		&manga.Status,
		&manga.TotalChapters,
		&manga.Description,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manga)
}
