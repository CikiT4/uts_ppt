package service

import (
	"errors"

	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"

	"github.com/google/uuid"
)

// ======================================================
// CHAT SERVICE
// ======================================================

type SendMessageRequest struct {
	Content     *string             `json:"content"`
	MessageType models.MessageType  `json:"message_type" binding:"required,oneof=text file image"`
}

type ChatService interface {
	SendMessage(consultationID uuid.UUID, senderID uuid.UUID, req *SendMessageRequest) (*models.ChatMessage, error)
	GetMessages(consultationID uuid.UUID, requesterID uuid.UUID, page, limit int) ([]*models.ChatMessage, int64, error)
	MarkRead(consultationID uuid.UUID, readerID uuid.UUID) error
}

type chatService struct {
	chatRepo    repository.ChatRepository
	consultRepo repository.ConsultationRepository
}

func NewChatService(chatRepo repository.ChatRepository, consultRepo repository.ConsultationRepository) ChatService {
	return &chatService{chatRepo: chatRepo, consultRepo: consultRepo}
}

func (s *chatService) SendMessage(consultationID uuid.UUID, senderID uuid.UUID, req *SendMessageRequest) (*models.ChatMessage, error) {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return nil, errors.New("consultation not found")
	}
	if c.ClientID != senderID && c.LawyerID != senderID {
		return nil, errors.New("access denied")
	}
	if c.Status == models.StatusCancelled || c.Status == models.StatusCompleted {
		return nil, errors.New("cannot send message to a closed consultation")
	}
	msg := &models.ChatMessage{
		ConsultationID: consultationID,
		SenderID:       senderID,
		MessageType:    req.MessageType,
		Content:        req.Content,
	}
	if err := s.chatRepo.SendMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *chatService) GetMessages(consultationID uuid.UUID, requesterID uuid.UUID, page, limit int) ([]*models.ChatMessage, int64, error) {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return nil, 0, errors.New("consultation not found")
	}
	if c.ClientID != requesterID && c.LawyerID != requesterID {
		return nil, 0, errors.New("access denied")
	}
	return s.chatRepo.GetMessages(consultationID, page, limit)
}

func (s *chatService) MarkRead(consultationID uuid.UUID, readerID uuid.UUID) error {
	return s.chatRepo.MarkRead(consultationID, readerID)
}

// ======================================================
// REVIEW SERVICE
// ======================================================

type CreateReviewRequest struct {
	Rating      int     `json:"rating" binding:"required,min=1,max=5"`
	Comment     *string `json:"comment"`
	IsAnonymous bool    `json:"is_anonymous"`
}

type ReviewService interface {
	Create(consultationID uuid.UUID, clientUserID uuid.UUID, req *CreateReviewRequest) (*models.Review, error)
	GetByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Review, int64, error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	consultRepo repository.ConsultationRepository
	lawyerRepo  repository.LawyerRepository
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	consultRepo repository.ConsultationRepository,
	lawyerRepo repository.LawyerRepository,
) ReviewService {
	return &reviewService{reviewRepo: reviewRepo, consultRepo: consultRepo, lawyerRepo: lawyerRepo}
}

func (s *reviewService) Create(consultationID uuid.UUID, clientUserID uuid.UUID, req *CreateReviewRequest) (*models.Review, error) {
	c, err := s.consultRepo.FindByID(consultationID)
	if err != nil || c == nil {
		return nil, errors.New("consultation not found")
	}
	if c.ClientID != clientUserID {
		return nil, errors.New("only the client can review this consultation")
	}
	if c.Status != models.StatusCompleted {
		return nil, errors.New("can only review completed consultations")
	}
	existing, _ := s.reviewRepo.FindByConsultationID(consultationID)
	if existing != nil {
		return nil, errors.New("review already submitted for this consultation")
	}
	review := &models.Review{
		ConsultationID: consultationID,
		ClientID:       clientUserID,
		LawyerID:       c.LawyerID,
		Rating:         req.Rating,
		Comment:        req.Comment,
		IsAnonymous:    req.IsAnonymous,
	}
	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *reviewService) GetByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Review, int64, error) {
	return s.reviewRepo.FindByLawyerID(lawyerID, page, limit)
}

// ======================================================
// SCHEDULE SERVICE
// ======================================================

type ScheduleRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

type ScheduleService interface {
	Create(lawyerUserID uuid.UUID, req *ScheduleRequest) (*models.LawyerSchedule, error)
	GetByLawyer(lawyerID uuid.UUID) ([]*models.LawyerSchedule, error)
	Delete(scheduleID uuid.UUID, lawyerUserID uuid.UUID) error
}

type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
	lawyerRepo   repository.LawyerRepository
}

func NewScheduleService(scheduleRepo repository.ScheduleRepository, lawyerRepo repository.LawyerRepository) ScheduleService {
	return &scheduleService{scheduleRepo: scheduleRepo, lawyerRepo: lawyerRepo}
}

func (s *scheduleService) Create(lawyerUserID uuid.UUID, req *ScheduleRequest) (*models.LawyerSchedule, error) {
	lawyer, err := s.lawyerRepo.FindByUserID(lawyerUserID)
	if err != nil || lawyer == nil {
		return nil, errors.New("lawyer profile not found")
	}
	schedule := &models.LawyerSchedule{
		LawyerID:  lawyer.ID,
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		IsActive:  true,
	}
	if err := s.scheduleRepo.Create(schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleService) GetByLawyer(lawyerID uuid.UUID) ([]*models.LawyerSchedule, error) {
	return s.scheduleRepo.FindByLawyerID(lawyerID)
}

func (s *scheduleService) Delete(scheduleID uuid.UUID, lawyerUserID uuid.UUID) error {
	return s.scheduleRepo.Delete(scheduleID)
}
