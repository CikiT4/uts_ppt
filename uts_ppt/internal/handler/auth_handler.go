package handler

import (
	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/service"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, "Registration successful", result)
}

// Login godoc
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.authService.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, "Login successful", result)
}

// GetProfile godoc
// GET /api/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Profile retrieved", user)
}

// UpdateProfile godoc
// PUT /api/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	user, err := h.authService.UpdateProfile(userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Profile updated", user)
}
