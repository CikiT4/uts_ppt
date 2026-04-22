package repository

import (
	"database/sql"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
)

type ReviewRepository interface {
	Create(r *models.Review) error
	FindByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Review, int64, error)
	FindByConsultationID(consultationID uuid.UUID) (*models.Review, error)
}

type reviewRepository struct{ db *sql.DB }

func NewReviewRepository(db *sql.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *models.Review) error {
	rv.ID = uuid.New()
	now := time.Now()
	rv.CreatedAt = now
	rv.UpdatedAt = now
	_, err := r.db.Exec(`INSERT INTO reviews
		(id, consultation_id, client_id, lawyer_id, rating, comment, is_anonymous, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rv.ID, rv.ConsultationID, rv.ClientID, rv.LawyerID,
		rv.Rating, rv.Comment, rv.IsAnonymous, rv.CreatedAt, rv.UpdatedAt)
	return err
}

func (r *reviewRepository) FindByLawyerID(lawyerID uuid.UUID, page, limit int) ([]*models.Review, int64, error) {
	var total int64
	r.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE lawyer_id=$1`, lawyerID).Scan(&total)
	offset := (page - 1) * limit
	rows, err := r.db.Query(`SELECT id, consultation_id, client_id, lawyer_id, rating,
		comment, is_anonymous, created_at, updated_at FROM reviews WHERE lawyer_id=$1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, lawyerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var reviews []*models.Review
	for rows.Next() {
		rv := &models.Review{}
		rows.Scan(&rv.ID, &rv.ConsultationID, &rv.ClientID, &rv.LawyerID,
			&rv.Rating, &rv.Comment, &rv.IsAnonymous, &rv.CreatedAt, &rv.UpdatedAt)
		reviews = append(reviews, rv)
	}
	return reviews, total, nil
}

func (r *reviewRepository) FindByConsultationID(consultationID uuid.UUID) (*models.Review, error) {
	rv := &models.Review{}
	err := r.db.QueryRow(`SELECT id, consultation_id, client_id, lawyer_id, rating,
		comment, is_anonymous, created_at, updated_at FROM reviews WHERE consultation_id=$1`,
		consultationID).Scan(&rv.ID, &rv.ConsultationID, &rv.ClientID, &rv.LawyerID,
		&rv.Rating, &rv.Comment, &rv.IsAnonymous, &rv.CreatedAt, &rv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return rv, err
}

// ======================================================
// SCHEDULE REPOSITORY
// ======================================================

type ScheduleRepository interface {
	Create(s *models.LawyerSchedule) error
	FindByLawyerID(lawyerID uuid.UUID) ([]*models.LawyerSchedule, error)
	Delete(id uuid.UUID) error
	Update(s *models.LawyerSchedule) error
}

type scheduleRepository struct{ db *sql.DB }

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(s *models.LawyerSchedule) error {
	s.ID = uuid.New()
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	_, err := r.db.Exec(`INSERT INTO lawyer_schedules
		(id, lawyer_id, day_of_week, start_time, end_time, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		s.ID, s.LawyerID, s.DayOfWeek, s.StartTime, s.EndTime, s.IsActive, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *scheduleRepository) FindByLawyerID(lawyerID uuid.UUID) ([]*models.LawyerSchedule, error) {
	rows, err := r.db.Query(`SELECT id, lawyer_id, day_of_week, start_time, end_time,
		is_active, created_at, updated_at FROM lawyer_schedules WHERE lawyer_id=$1 AND is_active=true
		ORDER BY day_of_week, start_time`, lawyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []*models.LawyerSchedule
	for rows.Next() {
		s := &models.LawyerSchedule{}
		rows.Scan(&s.ID, &s.LawyerID, &s.DayOfWeek, &s.StartTime, &s.EndTime,
			&s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func (r *scheduleRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE lawyer_schedules SET is_active=false, updated_at=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (r *scheduleRepository) Update(s *models.LawyerSchedule) error {
	_, err := r.db.Exec(`UPDATE lawyer_schedules SET day_of_week=$1, start_time=$2,
		end_time=$3, updated_at=$4 WHERE id=$5`,
		s.DayOfWeek, s.StartTime, s.EndTime, time.Now(), s.ID)
	return err
}
