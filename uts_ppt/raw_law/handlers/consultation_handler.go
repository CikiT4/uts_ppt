package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"raw-law-api/middleware"
	"raw-law-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BookConsultation godoc
// POST /api/consultations  [client only]
func BookConsultation(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB unavailable"})
		return
	}

	type Req struct {
		LawyerID        uint    `json:"lawyer_id" binding:"required"`
		ScheduleDate    string  `json:"schedule_date" binding:"required"`
		StartTime       string  `json:"start_time" binding:"required"`
		EndTime         string  `json:"end_time" binding:"required"`
		DurationHours   float64 `json:"duration_hours" binding:"required,min=0.5"`
		CaseDescription string  `json:"case_description" binding:"required,min=20"`
		CaseType        *string `json:"case_type"`
		Platform        string  `json:"platform" binding:"required,oneof=chat video_call in_person"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Validation failed", "errors": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	if userRole != "client" {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only clients can book a consultation"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Fetch lawyer for fee calculation
	var lawyer models.Lawyer
	if err := db.WithContext(ctx).First(&lawyer, req.LawyerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Lawyer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if !lawyer.IsAvailable {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Lawyer is not available"})
		return
	}

	fee := lawyer.ConsultationFeePerHour * req.DurationHours

	consultation := models.Consultation{
		ClientID:        userID.(uint),
		LawyerID:        req.LawyerID,
		ScheduleDate:    req.ScheduleDate,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		DurationHours:   req.DurationHours,
		Status:          models.StatusPending,
		CaseDescription: req.CaseDescription,
		CaseType:        req.CaseType,
		ConsultationFee: fee,
		Platform:        req.Platform,
	}

	// Use transaction: create consultation + payment atomically
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&consultation).Error; err != nil {
			return err
		}
		payment := models.Payment{
			ConsultationID: consultation.ID,
			Amount:         fee,
			PaymentStatus:  models.PaymentPending,
		}
		return tx.Create(&payment).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to book consultation",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Consultation booked successfully",
		"data":    consultation,
	})
}

// GetMyConsultations godoc
// GET /api/consultations?page=1&limit=10
func GetMyConsultations(c *gin.Context) {
	db := middleware.GetDB(c)
	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var consultations []models.Consultation
	var total int64

	query := db.WithContext(ctx).Model(&models.Consultation{}).
		Preload("Client").
		Preload("Lawyer").
		Preload("Lawyer.User")

	if role == "lawyer" {
		// Find lawyer profile first
		var lawyer models.Lawyer
		if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&lawyer).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Lawyer profile not found"})
			return
		}
		query = query.Where("lawyer_id = ?", lawyer.ID)
	} else {
		query = query.Where("client_id = ?", userID)
	}

	query.Count(&total)
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&consultations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Consultations retrieved",
		"data": gin.H{
			"items": consultations,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetConsultationStatus godoc
// GET /api/consultations/:id/status
func GetConsultationStatus(c *gin.Context) {
	db := middleware.GetDB(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var consultation models.Consultation
	if err := db.WithContext(ctx).First(&consultation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Consultation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status retrieved",
		"data": gin.H{
			"consultation_id": consultation.ID,
			"status":          consultation.Status,
			"updated_at":      consultation.UpdatedAt,
		},
	})
}

// CancelConsultation godoc
// PATCH /api/consultations/:id/cancel
func CancelConsultation(c *gin.Context) {
	db := middleware.GetDB(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid ID"})
		return
	}

	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Reason is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var consultation models.Consultation
	if err := db.WithContext(ctx).First(&consultation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Consultation not found"})
		return
	}
	if consultation.Status == models.StatusCompleted || consultation.Status == models.StatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot cancel this consultation"})
		return
	}

	db.WithContext(ctx).Model(&consultation).Updates(map[string]interface{}{
		"status":           models.StatusCancelled,
		"cancelled_reason": body.Reason,
	})

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Consultation cancelled"})
}
