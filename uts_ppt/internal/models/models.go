package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// USER MODEL
// ============================================================

type UserRole string

const (
	RoleClient UserRole = "client"
	RoleLawyer UserRole = "lawyer"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	Role            UserRole   `json:"role" db:"role"`
	FullName        string     `json:"full_name" db:"full_name"`
	Phone           *string    `json:"phone" db:"phone"`
	ProfilePhotoURL *string    `json:"profile_photo_url" db:"profile_photo_url"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// ============================================================
// LAWYER MODEL
// ============================================================

type Lawyer struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	UserID                 uuid.UUID `json:"user_id" db:"user_id"`
	LicenseNumber          string    `json:"license_number" db:"license_number"`
	Specialization         []string  `json:"specialization" db:"specialization"`
	YearsOfExperience      int       `json:"years_of_experience" db:"years_of_experience"`
	Education              *string   `json:"education" db:"education"`
	Bio                    *string   `json:"bio" db:"bio"`
	OfficeAddress          *string   `json:"office_address" db:"office_address"`
	City                   *string   `json:"city" db:"city"`
	Province               *string   `json:"province" db:"province"`
	Latitude               *float64  `json:"latitude" db:"latitude"`
	Longitude              *float64  `json:"longitude" db:"longitude"`
	ConsultationFeePerHour float64   `json:"consultation_fee_per_hour" db:"consultation_fee_per_hour"`
	IsAvailable            bool      `json:"is_available" db:"is_available"`
	Rating                 float64   `json:"rating" db:"rating"`
	TotalReviews           int       `json:"total_reviews" db:"total_reviews"`
	TotalConsultations     int       `json:"total_consultations" db:"total_consultations"`
	IsVerified             bool      `json:"is_verified" db:"is_verified"`
	VerificationDocumentURL *string  `json:"verification_document_url" db:"verification_document_url"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// LawyerDetail includes user info joined
type LawyerDetail struct {
	Lawyer
	User User `json:"user"`
}

// ============================================================
// CLIENT MODEL
// ============================================================

type Client struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	DateOfBirth *time.Time `json:"date_of_birth" db:"date_of_birth"`
	Address     *string    `json:"address" db:"address"`
	City        *string    `json:"city" db:"city"`
	Province    *string    `json:"province" db:"province"`
	Occupation  *string    `json:"occupation" db:"occupation"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ============================================================
// SCHEDULE MODEL
// ============================================================

type LawyerSchedule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	LawyerID    uuid.UUID `json:"lawyer_id" db:"lawyer_id"`
	DayOfWeek   int       `json:"day_of_week" db:"day_of_week"`
	StartTime   string    `json:"start_time" db:"start_time"`
	EndTime     string    `json:"end_time" db:"end_time"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ============================================================
// CONSULTATION MODEL
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
	ID               uuid.UUID          `json:"id" db:"id"`
	ClientID         uuid.UUID          `json:"client_id" db:"client_id"`
	LawyerID         uuid.UUID          `json:"lawyer_id" db:"lawyer_id"`
	ScheduleDate     string             `json:"schedule_date" db:"schedule_date"`
	StartTime        string             `json:"start_time" db:"start_time"`
	EndTime          string             `json:"end_time" db:"end_time"`
	DurationHours    float64            `json:"duration_hours" db:"duration_hours"`
	Status           ConsultationStatus `json:"status" db:"status"`
	CaseDescription  string             `json:"case_description" db:"case_description"`
	CaseType         *string            `json:"case_type" db:"case_type"`
	ConsultationFee  float64            `json:"consultation_fee" db:"consultation_fee"`
	Platform         string             `json:"platform" db:"platform"`
	MeetingLink      *string            `json:"meeting_link" db:"meeting_link"`
	Notes            *string            `json:"notes" db:"notes"`
	CancelledReason  *string            `json:"cancelled_reason" db:"cancelled_reason"`
	CancelledBy      *uuid.UUID         `json:"cancelled_by" db:"cancelled_by"`
	ConfirmedAt      *time.Time         `json:"confirmed_at" db:"confirmed_at"`
	CompletedAt      *time.Time         `json:"completed_at" db:"completed_at"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
}

// ============================================================
// PAYMENT MODEL
// ============================================================

type PaymentStatus string
type PaymentMethod string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentUploaded PaymentStatus = "uploaded"
	PaymentVerified PaymentStatus = "verified"
	PaymentRejected PaymentStatus = "rejected"
	PaymentRefunded PaymentStatus = "refunded"
)

const (
	MethodBankTransfer PaymentMethod = "bank_transfer"
	MethodEWallet      PaymentMethod = "e_wallet"
	MethodCreditCard   PaymentMethod = "credit_card"
)

type Payment struct {
	ID                 uuid.UUID      `json:"id" db:"id"`
	ConsultationID     uuid.UUID      `json:"consultation_id" db:"consultation_id"`
	Amount             float64        `json:"amount" db:"amount"`
	PaymentMethod      *PaymentMethod `json:"payment_method" db:"payment_method"`
	PaymentStatus      PaymentStatus  `json:"payment_status" db:"payment_status"`
	BankName           *string        `json:"bank_name" db:"bank_name"`
	AccountNumber      *string        `json:"account_number" db:"account_number"`
	TransferReference  *string        `json:"transfer_reference" db:"transfer_reference"`
	PaymentProofURL    *string        `json:"payment_proof_url" db:"payment_proof_url"`
	PaymentDate        *time.Time     `json:"payment_date" db:"payment_date"`
	VerifiedBy         *uuid.UUID     `json:"verified_by" db:"verified_by"`
	VerifiedAt         *time.Time     `json:"verified_at" db:"verified_at"`
	Notes              *string        `json:"notes" db:"notes"`
	CreatedAt          time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
}

// ============================================================
// REVIEW MODEL
// ============================================================

type Review struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ConsultationID uuid.UUID `json:"consultation_id" db:"consultation_id"`
	ClientID       uuid.UUID `json:"client_id" db:"client_id"`
	LawyerID       uuid.UUID `json:"lawyer_id" db:"lawyer_id"`
	Rating         int       `json:"rating" db:"rating"`
	Comment        *string   `json:"comment" db:"comment"`
	IsAnonymous    bool      `json:"is_anonymous" db:"is_anonymous"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ============================================================
// DOCUMENT MODEL
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
	ID             uuid.UUID    `json:"id" db:"id"`
	UploaderID     uuid.UUID    `json:"uploader_id" db:"uploader_id"`
	ConsultationID *uuid.UUID   `json:"consultation_id" db:"consultation_id"`
	DocumentType   DocumentType `json:"document_type" db:"document_type"`
	OriginalName   string       `json:"original_name" db:"original_name"`
	FileName       string       `json:"file_name" db:"file_name"`
	FilePath       string       `json:"-" db:"file_path"`
	FileURL        string       `json:"file_url" db:"file_url"`
	FileSize       int          `json:"file_size" db:"file_size"`
	MimeType       string       `json:"mime_type" db:"mime_type"`
	IsActive       bool         `json:"is_active" db:"is_active"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

// ============================================================
// CHAT MESSAGE MODEL
// ============================================================

type MessageType string

const (
	MessageText  MessageType = "text"
	MessageFile  MessageType = "file"
	MessageImage MessageType = "image"
)

type ChatMessage struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	ConsultationID uuid.UUID   `json:"consultation_id" db:"consultation_id"`
	SenderID       uuid.UUID   `json:"sender_id" db:"sender_id"`
	MessageType    MessageType `json:"message_type" db:"message_type"`
	Content        *string     `json:"content" db:"content"`
	FileURL        *string     `json:"file_url" db:"file_url"`
	FileName       *string     `json:"file_name" db:"file_name"`
	IsRead         bool        `json:"is_read" db:"is_read"`
	ReadAt         *time.Time  `json:"read_at" db:"read_at"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// ============================================================
// NOTIFICATION MODEL
// ============================================================

type Notification struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Title         string     `json:"title" db:"title"`
	Body          string     `json:"body" db:"body"`
	Type          string     `json:"type" db:"type"`
	ReferenceID   *uuid.UUID `json:"reference_id" db:"reference_id"`
	ReferenceType *string    `json:"reference_type" db:"reference_type"`
	IsRead        bool       `json:"is_read" db:"is_read"`
	ReadAt        *time.Time `json:"read_at" db:"read_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}
