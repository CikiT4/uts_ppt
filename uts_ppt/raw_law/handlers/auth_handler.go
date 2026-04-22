package handlers

import (
	"context"
	"net/http"
	"time"

	"raw-law-api/middleware"
	"raw-law-api/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ── Request DTOs ─────────────────────────────────────────────

type RegisterRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	FullName string          `json:"full_name" binding:"required"`
	Phone    string          `json:"phone"`
	Role     models.UserRole `json:"role" binding:"required,oneof=client lawyer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ── Handlers ─────────────────────────────────────────────────

// Register godoc
// POST /api/auth/register
func Register(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB unavailable"})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := err.Error()
		if errMsg == "EOF" {
			errMsg = "Body JSON kosong atau format tidak valid. Pastikan menggunakan format raw -> JSON dan diapit tanda kurung kurawal {}"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Validation failed",
			"errors":  errMsg,
		})
		return
	}

	// Check duplicate email with context timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var existing models.User
	if err := db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to hash password"})
		return
	}

	phone := req.Phone
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		FullName:     req.FullName,
		Phone:        &phone,
		IsActive:     true,
	}

	if err := db.WithContext(ctx).Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registration successful",
		"data": gin.H{
			"user":         user,
			"access_token": token,
		},
	})
}

// Login godoc
// POST /api/auth/login
func Login(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB unavailable"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Validation failed",
			"errors":  err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := db.WithContext(ctx).Where("email = ? AND is_active = true", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Query error", "error": err.Error()})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid email or password"})
		return
	}

	// Update last login
	now := time.Now()
	db.WithContext(ctx).Model(&user).Update("last_login_at", now)

	token, err := middleware.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"user":         user,
			"access_token": token,
		},
	})
}

// GetProfile godoc
// GET /api/profile
func GetProfile(c *gin.Context) {
	db := middleware.GetDB(c)
	userID, _ := c.Get("user_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := db.WithContext(ctx).First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Profile retrieved", "data": user})
}
