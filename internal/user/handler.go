package user

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"mangahub/internal/auth"
	"mangahub/internal/tcp"
	"mangahub/internal/udp"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB        *sql.DB
	TCPServer *tcp.Server
	UDPServer *udp.Server
}

func NewHandler(db *sql.DB, tcpServer *tcp.Server, udpServer *udp.Server) *Handler {
	return &Handler{
		DB:        db,
		TCPServer: tcpServer,
		UDPServer: udpServer,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	protected := r.Group("/users")
	protected.Use(auth.AuthMiddleware())

	protected.POST("/library", h.AddToLibrary)
	protected.GET("/library", h.GetLibrary)
	protected.PUT("/progress", h.UpdateProgress)
}

func (h *Handler) AddToLibrary(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MangaID        string `json:"manga_id"`
		Status         string `json:"status"`
		CurrentChapter int    `json:"current_chapter"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	req.MangaID = strings.TrimSpace(req.MangaID)
	req.Status = strings.TrimSpace(req.Status)

	if req.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}

	if req.Status == "" {
		req.Status = "reading"
	}

	if !isValidReadingStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be reading, completed, or plan_to_read",
		})
		return
	}

	if req.CurrentChapter < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "current_chapter must be greater than or equal to 0",
		})
		return
	}

	var totalChapters int

	err := h.DB.QueryRow(`
		SELECT total_chapters
		FROM manga
		WHERE id = ?
	`, req.MangaID).Scan(&totalChapters)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "manga not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if totalChapters > 0 && req.CurrentChapter > totalChapters {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "current_chapter exceeds total chapters",
			"total_chapters": totalChapters,
		})
		return
	}

	_, err = h.DB.Exec(`
		INSERT OR REPLACE INTO user_progress 
		(user_id, manga_id, current_chapter, status) 
		VALUES (?, ?, ?, ?)
	`, userID, req.MangaID, req.CurrentChapter, req.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "manga added to library",
		"manga_id":        req.MangaID,
		"status":          req.Status,
		"current_chapter": req.CurrentChapter,
	})
}

func (h *Handler) GetLibrary(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.DB.Query(`
		SELECT 
			m.id,
			m.title,
			m.author,
			up.current_chapter,
			up.status,
			up.updated_at
		FROM user_progress up
		JOIN manga m ON up.manga_id = m.id
		WHERE up.user_id = ?
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	library := []gin.H{}

	for rows.Next() {
		var id, title, author, status, updatedAt string
		var currentChapter int

		if err := rows.Scan(&id, &title, &author, &currentChapter, &status, &updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		library = append(library, gin.H{
			"id":              id,
			"title":           title,
			"author":          author,
			"current_chapter": currentChapter,
			"status":          status,
			"updated_at":      updatedAt,
		})
	}

	c.JSON(http.StatusOK, library)
}
func (h *Handler) UpdateProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req struct {
		MangaID        string `json:"manga_id" binding:"required"`
		CurrentChapter int    `json:"current_chapter" binding:"required"`
		Status         string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}
	req.MangaID = strings.TrimSpace(req.MangaID)
	req.Status = strings.TrimSpace(req.Status)

	if req.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "manga_id is required",
		})
		return
	}

	if req.Status == "" {
		req.Status = "reading"
	}

	if !isValidReadingStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be reading, completed, or plan_to_read",
		})
		return
	}

	if req.CurrentChapter <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "current_chapter must be greater than 0",
		})
		return
	}

	var totalChapters int

	err := h.DB.QueryRow(`
		SELECT total_chapters
		FROM manga
		WHERE id = ?
	`, req.MangaID).Scan(&totalChapters)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "manga not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if totalChapters > 0 && req.CurrentChapter > totalChapters {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "current_chapter exceeds total chapters",
			"total_chapters": totalChapters,
		})
		return
	}

	var exists int

	err = h.DB.QueryRow(`
		SELECT COUNT(*)
		FROM user_progress
		WHERE user_id = ? AND manga_id = ?
	`, userID, req.MangaID).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if exists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "manga is not in your library. Please add it first",
		})
		return
	}

	result, err := h.DB.Exec(`
		UPDATE user_progress
		SET current_chapter = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND manga_id = ?
		`, req.CurrentChapter, req.Status, userID, req.MangaID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "progress was not updated",
		})
		return
	}

	h.TCPServer.BroadcastCh <- tcp.NewProgressUpdate(userID, req.MangaID, req.CurrentChapter)

	if h.UDPServer != nil {
		h.UDPServer.Broadcast(udp.Notification{
			Type:      "progress_update",
			MangaID:   req.MangaID,
			Message:   "User updated reading progress",
			Timestamp: time.Now().Unix(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "progress updated",
		"manga_id":        req.MangaID,
		"current_chapter": req.CurrentChapter,
		"status":          req.Status,
	})
}
func isValidReadingStatus(status string) bool {
	return status == "reading" ||
		status == "completed" ||
		status == "plan_to_read"
}
