package notification

import (
	"net/http"
	"time"

	"mangahub/internal/udp"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	UDPServer *udp.Server
}

func NewHandler(udpServer *udp.Server) *Handler {
	return &Handler{
		UDPServer: udpServer,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/notify/chapter", h.NotifyChapterRelease)
}

func (h *Handler) NotifyChapterRelease(c *gin.Context) {
	var req struct {
		MangaID string `json:"manga_id" binding:"required"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input",
		})
		return
	}

	if req.Message == "" {
		req.Message = "New chapter released"
	}

	notification := udp.Notification{
		Type:      "chapter_release",
		MangaID:   req.MangaID,
		Message:   req.Message,
		Timestamp: time.Now().Unix(),
	}

	h.UDPServer.Broadcast(notification)

	c.JSON(http.StatusOK, gin.H{
		"message":      "chapter release notification sent",
		"manga_id":     req.MangaID,
		"notification": notification,
	})
}
