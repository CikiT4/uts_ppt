package service

import (
	"errors"

	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"

	"github.com/google/uuid"
)

type LawyerProfileRequest struct {
	LicenseNumber          string   `json:"license_number" binding:"required"`
	Specialization         []string `json:"specialization" binding:"required,min=1"`
	YearsOfExperience      int      `json:"years_of_experience" binding:"required,min=0"`
	Education              *string  `json:"education"`
	Bio                    *string  `json:"bio"`
	OfficeAddress          *string  `json:"office_address"`
	City                   *string  `json:"city"`
	Province               *string  `json:"province"`
	Latitude               *float64 `json:"latitude"`
	Longitude              *float64 `json:"longitude"`
	ConsultationFeePerHour float64  `json:"consultation_fee_per_hour" binding:"required,min=0"`
}

type LawyerService interface {
	CreateProfile(userID uuid.UUID, req *LawyerProfileRequest) (*models.Lawyer, error)
	GetProfile(lawyerID uuid.UUID) (*models.Lawyer, error)
	GetByUserID(userID uuid.UUID) (*models.Lawyer, error)
	UpdateProfile(lawyerID uuid.UUID, req *LawyerProfileRequest) (*models.Lawyer, error)
	SearchLawyers(filter repository.LawyerFilter) ([]*models.Lawyer, int64, error)
	SetAvailability(lawyerID uuid.UUID, available bool) error
}

type lawyerService struct {
	lawyerRepo repository.LawyerRepository
	userRepo   repository.UserRepository
}

func NewLawyerService(lawyerRepo repository.LawyerRepository, userRepo repository.UserRepository) LawyerService {
	return &lawyerService{lawyerRepo: lawyerRepo, userRepo: userRepo}
}

func (s *lawyerService) CreateProfile(userID uuid.UUID, req *LawyerProfileRequest) (*models.Lawyer, error) {
	existing, err := s.lawyerRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("lawyer profile already exists")
	}

	lawyer := &models.Lawyer{
		UserID:                 userID,
		LicenseNumber:          req.LicenseNumber,
		Specialization:         req.Specialization,
		YearsOfExperience:      req.YearsOfExperience,
		Education:              req.Education,
		Bio:                    req.Bio,
		OfficeAddress:          req.OfficeAddress,
		City:                   req.City,
		Province:               req.Province,
		Latitude:               req.Latitude,
		Longitude:              req.Longitude,
		ConsultationFeePerHour: req.ConsultationFeePerHour,
		IsAvailable:            true,
	}

	if err := s.lawyerRepo.Create(lawyer); err != nil {
		return nil, err
	}
	return lawyer, nil
}

func (s *lawyerService) GetProfile(lawyerID uuid.UUID) (*models.Lawyer, error) {
	lawyer, err := s.lawyerRepo.FindByID(lawyerID)
	if err != nil {
		return nil, err
	}
	if lawyer == nil {
		return nil, errors.New("lawyer not found")
	}
	return lawyer, nil
}

func (s *lawyerService) GetByUserID(userID uuid.UUID) (*models.Lawyer, error) {
	lawyer, err := s.lawyerRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if lawyer == nil {
		return nil, errors.New("lawyer profile not found")
	}
	return lawyer, nil
}

func (s *lawyerService) UpdateProfile(lawyerID uuid.UUID, req *LawyerProfileRequest) (*models.Lawyer, error) {
	lawyer, err := s.lawyerRepo.FindByID(lawyerID)
	if err != nil {
		return nil, err
	}
	if lawyer == nil {
		return nil, errors.New("lawyer not found")
	}

	lawyer.Specialization = req.Specialization
	lawyer.YearsOfExperience = req.YearsOfExperience
	lawyer.Education = req.Education
	lawyer.Bio = req.Bio
	lawyer.OfficeAddress = req.OfficeAddress
	lawyer.City = req.City
	lawyer.Province = req.Province
	lawyer.Latitude = req.Latitude
	lawyer.Longitude = req.Longitude
	lawyer.ConsultationFeePerHour = req.ConsultationFeePerHour

	if err := s.lawyerRepo.Update(lawyer); err != nil {
		return nil, err
	}
	return lawyer, nil
}

func (s *lawyerService) SearchLawyers(filter repository.LawyerFilter) ([]*models.Lawyer, int64, error) {
	return s.lawyerRepo.FindAll(filter)
}

func (s *lawyerService) SetAvailability(lawyerID uuid.UUID, available bool) error {
	return s.lawyerRepo.UpdateAvailability(lawyerID, available)
}
