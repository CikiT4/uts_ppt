package repository

import (
	"database/sql"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
)

type ConsultationRepository interface {
	Create(c *models.Consultation) error
	FindByID(id uuid.UUID) (*models.Consultation, error)
	FindByClientID(clientID uuid.UUID, page, limit int) ([]*models.Consultation, int64, error)
	FindByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Consultation, int64, error)
	UpdateStatus(id uuid.UUID, status models.ConsultationStatus) error
	Cancel(id uuid.UUID, reason string, cancelledBy uuid.UUID) error
	Complete(id uuid.UUID) error
	Confirm(id uuid.UUID) error
}

type consultationRepository struct{ db *sql.DB }

func NewConsultationRepository(db *sql.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) Create(c *models.Consultation) error {
	c.ID = uuid.New()
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	_, err := r.db.Exec(`
		INSERT INTO consultations
		(id, client_id, lawyer_id, schedule_date, start_time, end_time, duration_hours,
		status, case_description, case_type, consultation_fee, platform, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		c.ID, c.ClientID, c.LawyerID, c.ScheduleDate, c.StartTime, c.EndTime,
		c.DurationHours, c.Status, c.CaseDescription, c.CaseType,
		c.ConsultationFee, c.Platform, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *consultationRepository) FindByID(id uuid.UUID) (*models.Consultation, error) {
	c := &models.Consultation{}
	err := r.db.QueryRow(`SELECT id, client_id, lawyer_id, schedule_date, start_time, end_time,
		duration_hours, status, case_description, case_type, consultation_fee, platform,
		meeting_link, notes, cancelled_reason, cancelled_by, confirmed_at, completed_at,
		created_at, updated_at FROM consultations WHERE id=$1`, id).Scan(
		&c.ID, &c.ClientID, &c.LawyerID, &c.ScheduleDate, &c.StartTime, &c.EndTime,
		&c.DurationHours, &c.Status, &c.CaseDescription, &c.CaseType, &c.ConsultationFee,
		&c.Platform, &c.MeetingLink, &c.Notes, &c.CancelledReason, &c.CancelledBy,
		&c.ConfirmedAt, &c.CompletedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *consultationRepository) FindByClientID(clientID uuid.UUID, page, limit int) ([]*models.Consultation, int64, error) {
	return r.findPaginated("client_id", clientID, page, limit)
}

func (r *consultationRepository) FindByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Consultation, int64, error) {
	return r.findPaginated("lawyer_id", lawyerID, page, limit)
}

func (r *consultationRepository) findPaginated(col string, id uuid.UUID, page, limit int) ([]*models.Consultation, int64, error) {
	var total int64
	r.db.QueryRow("SELECT COUNT(*) FROM consultations WHERE "+col+"=$1", id).Scan(&total)
	offset := (page - 1) * limit
	rows, err := r.db.Query(`SELECT id, client_id, lawyer_id, schedule_date, start_time, end_time,
		duration_hours, status, case_description, case_type, consultation_fee, platform,
		meeting_link, notes, cancelled_reason, cancelled_by, confirmed_at, completed_at,
		created_at, updated_at FROM consultations WHERE `+col+`=$1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, id, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*models.Consultation
	for rows.Next() {
		c := &models.Consultation{}
		rows.Scan(&c.ID, &c.ClientID, &c.LawyerID, &c.ScheduleDate, &c.StartTime, &c.EndTime,
			&c.DurationHours, &c.Status, &c.CaseDescription, &c.CaseType, &c.ConsultationFee,
			&c.Platform, &c.MeetingLink, &c.Notes, &c.CancelledReason, &c.CancelledBy,
			&c.ConfirmedAt, &c.CompletedAt, &c.CreatedAt, &c.UpdatedAt)
		list = append(list, c)
	}
	return list, total, nil
}

func (r *consultationRepository) UpdateStatus(id uuid.UUID, status models.ConsultationStatus) error {
	_, err := r.db.Exec(`UPDATE consultations SET status=$1, updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	return err
}

func (r *consultationRepository) Cancel(id uuid.UUID, reason string, cancelledBy uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE consultations SET status='cancelled', cancelled_reason=$1,
		cancelled_by=$2, updated_at=$3 WHERE id=$4`, reason, cancelledBy, time.Now(), id)
	return err
}

func (r *consultationRepository) Complete(id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE consultations SET status='completed', completed_at=$1, updated_at=$1 WHERE id=$2`, now, id)
	return err
}

func (r *consultationRepository) Confirm(id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE consultations SET status='confirmed', confirmed_at=$1, updated_at=$1 WHERE id=$2`, now, id)
	return err
}
