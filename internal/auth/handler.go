package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
}

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Xóa khoảng trắng đầu/cuối
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	// Validate username/email/password rỗng
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	// Validate đơn giản cho password
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters"})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()

	_, err = h.DB.Exec(`
		INSERT INTO users (id, username, email, password_hash)
		VALUES (?, ?, ?, ?)
	`, id, req.Username, req.Email, hash)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "user created"})
}
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	var id, hash string

	err := h.DB.QueryRow(`
		SELECT id, password_hash FROM users WHERE username = ?
	`, req.Username).Scan(&id, &hash)

	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	if !CheckPassword(hash, req.Password) {
		c.JSON(401, gin.H{"error": "wrong password"})
		return
	}

	token, err := GenerateToken(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"token": token})
}
