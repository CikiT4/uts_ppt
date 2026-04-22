package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"raw-law-api/middleware"
	"raw-law-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListLawyers godoc
// GET /api/lawyers?city=Jakarta&specialization=Pidana&min_rating=4&max_fee=500000&page=1&limit=10
func ListLawyers(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB unavailable"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := db.WithContext(ctx).Model(&models.Lawyer{}).
		Preload("User").
		Where("lawyers.is_available = true")

	if city := c.Query("city"); city != "" {
		query = query.Where("LOWER(city) LIKE ?", "%"+strings.ToLower(city)+"%")
	}
	if spec := c.Query("specialization"); spec != "" {
		query = query.Where("FIND_IN_SET(?, specialization) > 0", spec)
	}
	if minRating := c.Query("min_rating"); minRating != "" {
		query = query.Where("rating >= ?", minRating)
	}
	if maxFee := c.Query("max_fee"); maxFee != "" {
		query = query.Where("consultation_fee_per_hour <= ?", maxFee)
	}
	if search := c.Query("search"); search != "" {
		query = query.Joins("JOIN users ON users.id = lawyers.user_id").
			Where("LOWER(users.full_name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Count query failed", "error": err.Error()})
		return
	}

	var lawyers []models.Lawyer
	if err := query.Order("rating DESC").Limit(limit).Offset(offset).Find(&lawyers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Query failed", "error": err.Error()})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Lawyers retrieved",
		"data": gin.H{
			"items":       lawyers,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

// GetLawyer godoc
// GET /api/lawyers/:id
func GetLawyer(c *gin.Context) {
	db := middleware.GetDB(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid lawyer ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var lawyer models.Lawyer
	if err := db.WithContext(ctx).Preload("User").First(&lawyer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Lawyer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lawyer retrieved", "data": lawyer})
}

// CreateLawyerProfile godoc
// POST /api/lawyers/profile  [lawyer only]
func CreateLawyerProfile(c *gin.Context) {
	db := middleware.GetDB(c)
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	if userRole != "lawyer" {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only lawyers can create a lawyer profile"})
		return
	}

	type Req struct {
		LicenseNumber          string   `json:"license_number" binding:"required"`
		Specialization         []string `json:"specialization" binding:"required,min=1"`
		YearsOfExperience      int      `json:"years_of_experience"`
		Education              *string  `json:"education"`
		Bio                    *string  `json:"bio"`
		OfficeAddress          *string  `json:"office_address"`
		City                   *string  `json:"city"`
		Province               *string  `json:"province"`
		ConsultationFeePerHour float64  `json:"consultation_fee_per_hour" binding:"required,min=0"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Validation failed", "errors": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check duplicate
	var existing models.Lawyer
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Lawyer profile already exists"})
		return
	}

	lawyer := models.Lawyer{
		UserID:                 userID.(uint),
		LicenseNumber:          req.LicenseNumber,
		Specialization:         strings.Join(req.Specialization, ","),
		YearsOfExperience:      req.YearsOfExperience,
		Education:              req.Education,
		Bio:                    req.Bio,
		OfficeAddress:          req.OfficeAddress,
		City:                   req.City,
		Province:               req.Province,
		ConsultationFeePerHour: req.ConsultationFeePerHour,
		IsAvailable:            true,
	}

	if err := db.WithContext(ctx).Create(&lawyer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create profile", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Lawyer profile created", "data": lawyer})
}

// GetLawyerReviews godoc
// GET /api/lawyers/:id/reviews
func GetLawyerReviews(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var reviews []models.Review
	var total int64

	db.WithContext(ctx).Model(&models.Review{}).Where("lawyer_id = ?", id).Count(&total)
	db.WithContext(ctx).Where("lawyer_id = ?", id).Limit(limit).Offset(offset).
		Order("created_at DESC").Find(&reviews)

	c.JSON(http.StatusOK, gin.H{
		"status": "success", "message": "Reviews retrieved",
		"data": gin.H{"items": reviews, "total": total, "page": page, "limit": limit},
	})
}
