package handler

import (
	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/service"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(ps service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: ps}
}

// GET /api/consultations/:id/payment
func (h *PaymentHandler) GetByConsultation(c *gin.Context) {
	consultationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid consultation ID", nil)
		return
	}
	payment, err := h.paymentService.GetByConsultationID(consultationID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Payment retrieved", payment)
}

// POST /api/payments/:id/upload — Upload payment proof (multipart form)
func (h *PaymentHandler) UploadProof(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payment ID", nil)
		return
	}

	file, err := c.FormFile("proof")
	if err != nil {
		response.BadRequest(c, "Payment proof file is required", err.Error())
		return
	}

	req := &service.UploadProofRequest{
		PaymentMethod: models.PaymentMethod(c.PostForm("payment_method")),
		BankName:      c.PostForm("bank_name"),
		Reference:     c.PostForm("transfer_reference"),
	}

	payment, err := h.paymentService.UploadProof(paymentID, file, req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Payment proof uploaded", payment)
}

// PATCH /api/payments/:id/verify  (Admin only)
func (h *PaymentHandler) Verify(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payment ID", nil)
		return
	}
	adminID := middleware.GetUserID(c)
	if err := h.paymentService.Verify(paymentID, adminID); err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}
	response.Success(c, "Payment verified", nil)
}

// PATCH /api/payments/:id/reject  (Admin only)
func (h *PaymentHandler) Reject(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payment ID", nil)
		return
	}
	var body struct {
		Notes string `json:"notes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Notes required", err.Error())
		return
	}
	if err := h.paymentService.Reject(paymentID, body.Notes); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Payment rejected", nil)
}
