package repository

import (
	"database/sql"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(p *models.Payment) error
	FindByConsultationID(consultationID uuid.UUID) (*models.Payment, error)
	FindByID(id uuid.UUID) (*models.Payment, error)
	UploadProof(id uuid.UUID, proofURL string, method models.PaymentMethod, bankName, reference string) error
	Verify(id uuid.UUID, verifiedBy uuid.UUID) error
	Reject(id uuid.UUID, notes string) error
}

type paymentRepository struct{ db *sql.DB }

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(p *models.Payment) error {
	p.ID = uuid.New()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err := r.db.Exec(`INSERT INTO payments
		(id, consultation_id, amount, payment_status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		p.ID, p.ConsultationID, p.Amount, p.PaymentStatus, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *paymentRepository) FindByConsultationID(consultationID uuid.UUID) (*models.Payment, error) {
	return r.scan(`SELECT id, consultation_id, amount, payment_method, payment_status,
		bank_name, account_number, transfer_reference, payment_proof_url,
		payment_date, verified_by, verified_at, notes, created_at, updated_at
		FROM payments WHERE consultation_id=$1`, consultationID)
}

func (r *paymentRepository) FindByID(id uuid.UUID) (*models.Payment, error) {
	return r.scan(`SELECT id, consultation_id, amount, payment_method, payment_status,
		bank_name, account_number, transfer_reference, payment_proof_url,
		payment_date, verified_by, verified_at, notes, created_at, updated_at
		FROM payments WHERE id=$1`, id)
}

func (r *paymentRepository) UploadProof(id uuid.UUID, proofURL string, method models.PaymentMethod, bankName, reference string) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE payments SET payment_proof_url=$1, payment_method=$2,
		bank_name=$3, transfer_reference=$4, payment_status='uploaded',
		payment_date=$5, updated_at=$5 WHERE id=$6`,
		proofURL, method, bankName, reference, now, id)
	return err
}

func (r *paymentRepository) Verify(id uuid.UUID, verifiedBy uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE payments SET payment_status='verified',
		verified_by=$1, verified_at=$2, updated_at=$2 WHERE id=$3`, verifiedBy, now, id)
	return err
}

func (r *paymentRepository) Reject(id uuid.UUID, notes string) error {
	_, err := r.db.Exec(`UPDATE payments SET payment_status='rejected', notes=$1, updated_at=$2 WHERE id=$3`,
		notes, time.Now(), id)
	return err
}

func (r *paymentRepository) scan(query string, args ...interface{}) (*models.Payment, error) {
	p := &models.Payment{}
	err := r.db.QueryRow(query, args...).Scan(
		&p.ID, &p.ConsultationID, &p.Amount, &p.PaymentMethod, &p.PaymentStatus,
		&p.BankName, &p.AccountNumber, &p.TransferReference, &p.PaymentProofURL,
		&p.PaymentDate, &p.VerifiedBy, &p.VerifiedAt, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}
