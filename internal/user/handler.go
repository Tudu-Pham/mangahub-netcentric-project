package user

import (
	"database/sql"
	"net/http"

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

	_, err := h.DB.Exec(`
		INSERT OR REPLACE INTO user_progress 
		(user_id, manga_id, current_chapter, status) 
		VALUES (?, ?, ?, ?)
	`, userID, req.MangaID, req.CurrentChapter, req.Status)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "manga added to library",
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

	var req struct {
		MangaID        string `json:"manga_id"`
		CurrentChapter int    `json:"current_chapter"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// update progress
	_, err := h.DB.Exec(`
		UPDATE user_progress
		SET current_chapter = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND manga_id = ?
	`, req.CurrentChapter, userID, req.MangaID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	h.TCPServer.BroadcastCh <- tcp.NewProgressUpdate(userID, req.MangaID, req.CurrentChapter)

	h.UDPServer.SendNotification(
		udp.Notification{
			Type:    "progress_update",
			Message: "User updated reading progress",
		},
		"127.0.0.1:9999", // client port
	)

	c.JSON(200, gin.H{
		"message": "progress updated",
	})
}
