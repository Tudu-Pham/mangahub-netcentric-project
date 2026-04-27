package manga

import (
	"database/sql"
	"net/http"
	"strconv"

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
	r.GET("/seed/manga", h.SeedManga)
}

func (h *Handler) GetAllManga(c *gin.Context) {
	q := c.Query("q")
	genre := c.Query("genre")
	status := c.Query("status")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	query := `
		SELECT id, title, author, genres, status, total_chapters, description
		FROM manga
		WHERE 1=1
	`

	args := []interface{}{}

	if q != "" {
		query += " AND (LOWER(title) LIKE LOWER(?) OR LOWER(author) LIKE LOWER(?))"
		search := "%" + q + "%"
		args = append(args, search, search)
	}

	if genre != "" {
		query += " AND LOWER(genres) LIKE LOWER(?)"
		args = append(args, "%"+genre+"%")
	}

	if status != "" {
		query += " AND LOWER(status) = LOWER(?)"
		args = append(args, status)
	}

	query += " ORDER BY title ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	mangas := []gin.H{}

	for rows.Next() {
		var id, title, author, genres, mangaStatus, description string
		var totalChapters int

		if err := rows.Scan(
			&id,
			&title,
			&author,
			&genres,
			&mangaStatus,
			&totalChapters,
			&description,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		mangas = append(mangas, gin.H{
			"id":             id,
			"title":          title,
			"author":         author,
			"genres":         genres,
			"status":         mangaStatus,
			"total_chapters": totalChapters,
			"description":    description,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"page":    page,
		"limit":   limit,
		"results": mangas,
	})
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

func (h *Handler) SeedManga(c *gin.Context) {
	err := SeedMangaFromJSON(h.DB, "data/manga.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "manga data seeded successfully",
	})
}
