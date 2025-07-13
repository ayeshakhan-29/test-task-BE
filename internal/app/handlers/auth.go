package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/ayeshakhan-29/test-task-BE/internal/app/models"
	"github.com/ayeshakhan-29/test-task-BE/internal/auth"
	"github.com/ayeshakhan-29/test-task-BE/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *database.Database
}

func NewAuthHandler(db *database.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var count int
	err := h.db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user with full name
	result, err := h.db.DB.Exec(`
		INSERT INTO users (full_name, email, password_hash)
		VALUES (?, ?, ?)
	`, req.FullName, req.Email, string(hashedPassword))

	if err != nil {
		// Log the actual error for debugging
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	userID, _ := result.LastInsertId()
	
	// Generate JWT token
	token, err := auth.GenerateToken(uint(userID), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user := &models.User{
		ID:        uint(userID),
		FullName:  req.FullName,
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.DB.QueryRow(`
		SELECT id, full_name, email, password_hash, created_at, updated_at 
		FROM users WHERE email = ?
	`, req.Email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.PasswordHash, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  &user,
	})
    }