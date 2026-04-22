package models

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================
// USER
// ============================================================

type UserRole string

const (
	RoleClient UserRole = "client"
	RoleLawyer UserRole = "lawyer"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Email           string         `json:"email" gorm:"uniqueIndex;size:255;not null"`
	PasswordHash    string         `json:"-" gorm:"column:password_hash;size:255;not null"`
	Role            UserRole       `json:"role" gorm:"type:enum('client','lawyer','admin');default:'client';not null"`
	FullName        string         `json:"full_name" gorm:"size:255;not null"`
	Phone           *string        `json:"phone" gorm:"size:20"`
	ProfilePhotoURL *string        `json:"profile_photo_url" gorm:"size:500"`
	IsActive        bool           `json:"is_active" gorm:"default:true;not null"`
	IsVerified      bool           `json:"is_verified" gorm:"default:false;not null"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================
// LAWYER
// ============================================================

type Lawyer struct {
	ID                     uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID                 uint           `json:"user_id" gorm:"uniqueIndex;not null"`
	User                   User           `json:"user" gorm:"foreignKey:UserID"`
	LicenseNumber          string         `json:"license_number" gorm:"uniqueIndex;size:100;not null"`
	Specialization         string         `json:"specialization" gorm:"size:500"`        // comma-separated
	YearsOfExperience      int            `json:"years_of_experience" gorm:"default:0"`
	Education              *string        `json:"education" gorm:"type:text"`
	Bio                    *string        `json:"bio" gorm:"type:text"`
	OfficeAddress          *string        `json:"office_address" gorm:"type:text"`
	City                   *string        `json:"city" gorm:"size:100"`
	Province               *string        `json:"province" gorm:"size:100"`
	ConsultationFeePerHour float64        `json:"consultation_fee_per_hour" gorm:"type:decimal(15,2);default:0"`
	IsAvailable            bool           `json:"is_available" gorm:"default:true;not null"`
	Rating                 float64        `json:"rating" gorm:"type:decimal(3,2);default:0"`
	TotalReviews           int            `json:"total_reviews" gorm:"default:0"`
	TotalConsultations     int            `json:"total_consultations" gorm:"default:0"`
	IsVerified             bool           `json:"is_verified" gorm:"default:false"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================
// CONSULTATION
// ============================================================

type ConsultationStatus string

const (
	StatusPending   ConsultationStatus = "pending"
	StatusConfirmed ConsultationStatus = "confirmed"
	StatusOngoing   ConsultationStatus = "ongoing"
	StatusCompleted ConsultationStatus = "completed"
	StatusCancelled ConsultationStatus = "cancelled"
)

type Consultation struct {
	ID              uint               `json:"id" gorm:"primaryKey;autoIncrement"`
	ClientID        uint               `json:"client_id" gorm:"not null;index"`
	Client          User               `json:"client" gorm:"foreignKey:ClientID"`
	LawyerID        uint               `json:"lawyer_id" gorm:"not null;index"`
	Lawyer          Lawyer             `json:"lawyer" gorm:"foreignKey:LawyerID"`
	ScheduleDate    string             `json:"schedule_date" gorm:"type:date;not null"`
	StartTime       string             `json:"start_time" gorm:"type:time;not null"`
	EndTime         string             `json:"end_time" gorm:"type:time;not null"`
	DurationHours   float64            `json:"duration_hours" gorm:"type:decimal(4,2);not null"`
	Status          ConsultationStatus `json:"status" gorm:"type:enum('pending','confirmed','ongoing','completed','cancelled');default:'pending'"`
	CaseDescription string             `json:"case_description" gorm:"type:text;not null"`
	CaseType        *string            `json:"case_type" gorm:"size:100"`
	ConsultationFee float64            `json:"consultation_fee" gorm:"type:decimal(15,2);not null"`
	Platform        string             `json:"platform" gorm:"size:50;default:'chat'"`
	Notes           *string            `json:"notes" gorm:"type:text"`
	CancelledReason *string            `json:"cancelled_reason" gorm:"type:text"`
	ConfirmedAt     *time.Time         `json:"confirmed_at"`
	CompletedAt     *time.Time         `json:"completed_at"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       gorm.DeletedAt     `json:"-" gorm:"index"`
}

// ============================================================
// PAYMENT
// ============================================================

type PaymentStatus string
type PaymentMethod string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentUploaded PaymentStatus = "uploaded"
	PaymentVerified PaymentStatus = "verified"
	PaymentRejected PaymentStatus = "rejected"
)

const (
	MethodBankTransfer PaymentMethod = "bank_transfer"
	MethodEWallet      PaymentMethod = "e_wallet"
	MethodCreditCard   PaymentMethod = "credit_card"
)

type Payment struct {
	ID                uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	ConsultationID    uint           `json:"consultation_id" gorm:"uniqueIndex;not null"`
	Consultation      Consultation   `json:"consultation" gorm:"foreignKey:ConsultationID"`
	Amount            float64        `json:"amount" gorm:"type:decimal(15,2);not null"`
	PaymentMethod     *PaymentMethod `json:"payment_method" gorm:"type:enum('bank_transfer','e_wallet','credit_card')"`
	PaymentStatus     PaymentStatus  `json:"payment_status" gorm:"type:enum('pending','uploaded','verified','rejected');default:'pending'"`
	BankName          *string        `json:"bank_name" gorm:"size:100"`
	AccountNumber     *string        `json:"account_number" gorm:"size:50"`
	TransferReference *string        `json:"transfer_reference" gorm:"size:100"`
	PaymentProofURL   *string        `json:"payment_proof_url" gorm:"size:500"`
	PaymentDate       *time.Time     `json:"payment_date"`
	VerifiedBy        *uint          `json:"verified_by"`
	VerifiedAt        *time.Time     `json:"verified_at"`
	Notes             *string        `json:"notes" gorm:"type:text"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// ============================================================
// REVIEW
// ============================================================

type Review struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	ConsultationID uint           `json:"consultation_id" gorm:"uniqueIndex;not null"`
	Consultation   Consultation   `json:"consultation" gorm:"foreignKey:ConsultationID"`
	ClientID       uint           `json:"client_id" gorm:"not null;index"`
	LawyerID       uint           `json:"lawyer_id" gorm:"not null;index"`
	Rating         int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment        *string        `json:"comment" gorm:"type:text"`
	IsAnonymous    bool           `json:"is_anonymous" gorm:"default:false"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================
// DOCUMENT
// ============================================================

type DocumentType string

const (
	DocConsultation DocumentType = "consultation_doc"
	DocPaymentProof DocumentType = "payment_proof"
	DocCaseFile     DocumentType = "case_file"
	DocProfilePhoto DocumentType = "profile_photo"
	DocLicense      DocumentType = "license"
)

type Document struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UploaderID     uint           `json:"uploader_id" gorm:"not null;index"`
	ConsultationID *uint          `json:"consultation_id" gorm:"index"`
	DocumentType   DocumentType   `json:"document_type" gorm:"type:enum('consultation_doc','payment_proof','case_file','profile_photo','license');not null"`
	OriginalName   string         `json:"original_name" gorm:"size:255;not null"`
	FileName       string         `json:"file_name" gorm:"size:255;not null"`
	FilePath       string         `json:"-" gorm:"size:500;not null"`
	FileURL        string         `json:"file_url" gorm:"size:500;not null"`
	FileSize       int64          `json:"file_size" gorm:"not null"`
	MimeType       string         `json:"mime_type" gorm:"size:100;not null"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
