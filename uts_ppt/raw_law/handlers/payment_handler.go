package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"raw-law-api/middleware"
	"raw-law-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const uploadDir = "./uploads/payments"

// UploadPaymentProof godoc
// POST /api/payments/:id/upload  (multipart/form-data)
// Fields: proof (file), payment_method, bank_name, transfer_reference
func UploadPaymentProof(c *gin.Context) {
	db := middleware.GetDB(c)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB unavailable"})
		return
	}

	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid payment ID"})
		return
	}

	// Validate required form fields
	paymentMethod := c.PostForm("payment_method")
	if paymentMethod == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "payment_method is required"})
		return
	}
	allowed := map[string]bool{"bank_transfer": true, "e_wallet": true, "credit_card": true}
	if !allowed[paymentMethod] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "payment_method must be one of: bank_transfer, e_wallet, credit_card",
		})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("proof")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "proof file is required"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Only jpg, jpeg, png, pdf files are allowed",
		})
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File size must not exceed 10MB"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Fetch payment record
	var payment models.Payment
	if err := db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Payment not found"})
		return
	}
	if payment.PaymentStatus == models.PaymentVerified {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Payment already verified"})
		return
	}

	// Save file to disk
	os.MkdirAll(uploadDir, 0755)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save file"})
		return
	}

	// Save document record
	fileURL := "/uploads/payments/" + fileName
	doc := models.Document{
		UploaderID:   payment.ConsultationID,
		DocumentType: models.DocPaymentProof,
		OriginalName: file.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileURL:      fileURL,
		FileSize:     file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		IsActive:     true,
	}
	db.WithContext(ctx).Create(&doc)

	// Update payment record
	method := models.PaymentMethod(paymentMethod)
	bankName := c.PostForm("bank_name")
	reference := c.PostForm("transfer_reference")
	now := time.Now()

	updates := map[string]interface{}{
		"payment_status":     models.PaymentUploaded,
		"payment_method":     &method,
		"payment_proof_url":  fileURL,
		"payment_date":       &now,
		"bank_name":          &bankName,
		"transfer_reference": &reference,
	}

	if err := db.WithContext(ctx).Model(&payment).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update payment", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment proof uploaded successfully",
		"data": gin.H{
			"payment_id":    payment.ID,
			"status":        models.PaymentUploaded,
			"proof_url":     fileURL,
			"uploaded_at":   now,
		},
	})
}

// GetPaymentByConsultation godoc
// GET /api/consultations/:id/payment
func GetPaymentByConsultation(c *gin.Context) {
	db := middleware.GetDB(c)
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var payment models.Payment
	if err := db.WithContext(ctx).Where("consultation_id = ?", id).First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Payment retrieved", "data": payment})
}
