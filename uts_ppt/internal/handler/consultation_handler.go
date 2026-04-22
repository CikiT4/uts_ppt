package handler

import (
	"strconv"

	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/service"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConsultationHandler struct {
	consultService service.ConsultationService
	reviewService  service.ReviewService
}

func NewConsultationHandler(cs service.ConsultationService, rs service.ReviewService) *ConsultationHandler {
	return &ConsultationHandler{consultService: cs, reviewService: rs}
}

// POST /api/consultations
func (h *ConsultationHandler) Book(c *gin.Context) {
	var req service.BookConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	consultation, err := h.consultService.Book(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Created(c, "Consultation booked successfully", consultation)
}

// GET /api/consultations
func (h *ConsultationHandler) GetMyConsultations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	consultations, total, err := h.consultService.GetMyConsultations(userID, role, page, limit)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Paginated(c, "Consultations retrieved", consultations, total, page, limit)
}

// GET /api/consultations/:id
func (h *ConsultationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	consultation, err := h.consultService.GetByID(id, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Consultation retrieved", consultation)
}

// PATCH /api/consultations/:id/confirm
func (h *ConsultationHandler) Confirm(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.consultService.Confirm(id, userID); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Consultation confirmed", nil)
}

// PATCH /api/consultations/:id/cancel
func (h *ConsultationHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	var req service.CancelConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Reason is required", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.consultService.Cancel(id, userID, req.Reason); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Consultation cancelled", nil)
}

// PATCH /api/consultations/:id/complete
func (h *ConsultationHandler) Complete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.consultService.Complete(id, userID); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Consultation completed", nil)
}

// POST /api/consultations/:id/reviews
func (h *ConsultationHandler) CreateReview(c *gin.Context) {
	consultationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	var req service.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	review, err := h.reviewService.Create(consultationID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Created(c, "Review submitted", review)
}

// GET /api/consultations/:id/status
func (h *ConsultationHandler) GetStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	consultation, err := h.consultService.GetByID(id, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Status retrieved", gin.H{
		"consultation_id": consultation.ID,
		"status":          consultation.Status,
		"updated_at":      consultation.UpdatedAt,
	})
}

// Ensure models is used
var _ = models.StatusPending
