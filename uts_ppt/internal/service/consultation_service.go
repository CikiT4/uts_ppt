package service

import (
	"errors"

	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"

	"github.com/google/uuid"
)

type BookConsultationRequest struct {
	LawyerID        string  `json:"lawyer_id" binding:"required,uuid"`
	ScheduleDate    string  `json:"schedule_date" binding:"required"`
	StartTime       string  `json:"start_time" binding:"required"`
	EndTime         string  `json:"end_time" binding:"required"`
	DurationHours   float64 `json:"duration_hours" binding:"required,min=0.5"`
	CaseDescription string  `json:"case_description" binding:"required,min=20"`
	CaseType        *string `json:"case_type"`
	Platform        string  `json:"platform" binding:"required,oneof=chat video_call in_person"`
}

type CancelConsultationRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ConsultationService interface {
	Book(clientUserID uuid.UUID, req *BookConsultationRequest) (*models.Consultation, error)
	GetByID(id uuid.UUID, requesterID uuid.UUID) (*models.Consultation, error)
	GetMyConsultations(userID uuid.UUID, role models.UserRole, page, limit int) ([]*models.Consultation, int64, error)
	Confirm(consultationID uuid.UUID, lawyerUserID uuid.UUID) error
	Cancel(consultationID uuid.UUID, userID uuid.UUID, reason string) error
	Complete(consultationID uuid.UUID, lawyerUserID uuid.UUID) error
}

type consultationService struct {
	consultRepo repository.ConsultationRepository
	lawyerRepo  repository.LawyerRepository
	paymentRepo repository.PaymentRepository
}

func NewConsultationService(
	consultRepo repository.ConsultationRepository,
	lawyerRepo repository.LawyerRepository,
	paymentRepo repository.PaymentRepository,
) ConsultationService {
	return &consultationService{consultRepo: consultRepo, lawyerRepo: lawyerRepo, paymentRepo: paymentRepo}
}

func (s *consultationService) Book(clientUserID uuid.UUID, req *BookConsultationRequest) (*models.Consultation, error) {
	lawyerID, _ := uuid.Parse(req.LawyerID)
	lawyer, err := s.lawyerRepo.FindByID(lawyerID)
	if err != nil || lawyer == nil {
		return nil, errors.New("lawyer not found")
	}
	if !lawyer.IsAvailable {
		return nil, errors.New("lawyer is not available")
	}

	fee := lawyer.ConsultationFeePerHour * req.DurationHours

	consultation := &models.Consultation{
		ClientID:        clientUserID,
		LawyerID:        lawyerID,
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

	if err := s.consultRepo.Create(consultation); err != nil {
		return nil, err
	}

	// Auto-create payment record
	payment := &models.Payment{
		ConsultationID: consultation.ID,
		Amount:         fee,
		PaymentStatus:  models.PaymentPending,
	}
	s.paymentRepo.Create(payment)

	return consultation, nil
}

func (s *consultationService) GetByID(id uuid.UUID, requesterID uuid.UUID) (*models.Consultation, error) {
	c, err := s.consultRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errors.New("consultation not found")
	}
	// Access control: only participants can view
	if c.ClientID != requesterID && c.LawyerID != requesterID {
		return nil, errors.New("access denied")
	}
	return c, nil
}

func (s *consultationService) GetMyConsultations(userID uuid.UUID, role models.UserRole, page, limit int) ([]*models.Consultation, int64, error) {
	if role == models.RoleLawyer {
		lawyer, err := s.lawyerRepo.FindByUserID(userID)
		if err != nil || lawyer == nil {
			return nil, 0, errors.New("lawyer profile not found")
		}
		return s.consultRepo.FindByLawyerID(lawyer.ID, page, limit)
	}
	return s.consultRepo.FindByClientID(userID, page, limit)
}

func (s *consultationService) Confirm(consultationID uuid.UUID, lawyerUserID uuid.UUID) error {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return errors.New("consultation not found")
	}
	lawyer, _ := s.lawyerRepo.FindByUserID(lawyerUserID)
	if lawyer == nil || c.LawyerID != lawyer.ID {
		return errors.New("access denied")
	}
	if c.Status != models.StatusPending {
		return errors.New("only pending consultations can be confirmed")
	}
	return s.consultRepo.Confirm(consultationID)
}

func (s *consultationService) Cancel(consultationID uuid.UUID, userID uuid.UUID, reason string) error {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return errors.New("consultation not found")
	}
	if c.Status == models.StatusCompleted || c.Status == models.StatusCancelled {
		return errors.New("cannot cancel this consultation")
	}
	return s.consultRepo.Cancel(consultationID, reason, userID)
}

func (s *consultationService) Complete(consultationID uuid.UUID, lawyerUserID uuid.UUID) error {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return errors.New("consultation not found")
	}
	lawyer, _ := s.lawyerRepo.FindByUserID(lawyerUserID)
	if lawyer == nil || c.LawyerID != lawyer.ID {
		return errors.New("access denied")
	}
	if c.Status != models.StatusConfirmed && c.Status != models.StatusOngoing {
		return errors.New("consultation must be confirmed before completing")
	}
	return s.consultRepo.Complete(consultationID)
}
