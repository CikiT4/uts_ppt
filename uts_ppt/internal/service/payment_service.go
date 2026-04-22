package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"legal-consultation-api/internal/config"
	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"

	"github.com/google/uuid"
)

type UploadProofRequest struct {
	PaymentMethod models.PaymentMethod `json:"payment_method" binding:"required"`
	BankName      string               `json:"bank_name"`
	Reference     string               `json:"transfer_reference"`
}

type PaymentService interface {
	GetByConsultationID(consultationID uuid.UUID) (*models.Payment, error)
	UploadProof(paymentID uuid.UUID, file *multipart.FileHeader, req *UploadProofRequest) (*models.Payment, error)
	Verify(paymentID uuid.UUID, adminID uuid.UUID) error
	Reject(paymentID uuid.UUID, notes string) error
}

type paymentService struct {
	paymentRepo     repository.PaymentRepository
	consultRepo     repository.ConsultationRepository
}

func NewPaymentService(paymentRepo repository.PaymentRepository, consultRepo repository.ConsultationRepository) PaymentService {
	return &paymentService{paymentRepo: paymentRepo, consultRepo: consultRepo}
}

func (s *paymentService) GetByConsultationID(consultationID uuid.UUID) (*models.Payment, error) {
	p, err := s.paymentRepo.FindByConsultationID(consultationID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("payment not found")
	}
	return p, nil
}

func (s *paymentService) UploadProof(paymentID uuid.UUID, file *multipart.FileHeader, req *UploadProofRequest) (*models.Payment, error) {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil || payment == nil {
		return nil, errors.New("payment not found")
	}
	if payment.PaymentStatus == models.PaymentVerified {
		return nil, errors.New("payment already verified")
	}

	// Save file
	uploadDir := config.AppConfig.UploadDir + "/payments"
	os.MkdirAll(uploadDir, 0755)

	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()
	buf := make([]byte, 1024*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	fileURL := "/uploads/payments/" + fileName
	if err := s.paymentRepo.UploadProof(paymentID, fileURL, req.PaymentMethod, req.BankName, req.Reference); err != nil {
		return nil, err
	}
	return s.paymentRepo.FindByID(paymentID)
}

func (s *paymentService) Verify(paymentID uuid.UUID, adminID uuid.UUID) error {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil || payment == nil {
		return errors.New("payment not found")
	}
	if payment.PaymentStatus != models.PaymentUploaded {
		return errors.New("payment proof must be uploaded first")
	}
	if err := s.paymentRepo.Verify(paymentID, adminID); err != nil {
		return err
	}
	// Confirm consultation
	s.consultRepo.Confirm(payment.ConsultationID)
	return nil
}

func (s *paymentService) Reject(paymentID uuid.UUID, notes string) error {
	return s.paymentRepo.Reject(paymentID, notes)
}

// Allowed types check
func isAllowedFileType(filename string) bool {
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	return allowed[strings.ToLower(filepath.Ext(filename))]
}
