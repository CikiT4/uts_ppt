package service

import (
	"errors"

	"legal-consultation-api/internal/models"
	"legal-consultation-api/internal/repository"
	jwtpkg "legal-consultation-api/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// DTO types
type RegisterRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	FullName string          `json:"full_name" binding:"required"`
	Phone    string          `json:"phone"`
	Role     models.UserRole `json:"role" binding:"required,oneof=client lawyer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User   *models.User       `json:"user"`
	Tokens *jwtpkg.TokenPair  `json:"tokens"`
}

type UpdateProfileRequest struct {
	FullName string  `json:"full_name" binding:"required"`
	Phone    *string `json:"phone"`
}

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*models.User, error)
}

type authService struct {
	userRepo   repository.UserRepository
	lawyerRepo repository.LawyerRepository
}

func NewAuthService(userRepo repository.UserRepository, lawyerRepo repository.LawyerRepository) AuthService {
	return &authService{userRepo: userRepo, lawyerRepo: lawyerRepo}
}

func (s *authService) Register(req *RegisterRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	phone := req.Phone
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		FullName:     req.FullName,
		Phone:        &phone,
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	s.userRepo.UpdateLastLogin(user.ID)

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *authService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *authService) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	user.FullName = req.FullName
	user.Phone = req.Phone
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}
