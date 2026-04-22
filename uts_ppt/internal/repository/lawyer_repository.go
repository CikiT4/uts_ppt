package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type LawyerFilter struct {
	Specialization string
	City           string
	MinRating      float64
	MaxFee         float64
	MinFee         float64
	Search         string
	Page           int
	Limit          int
}

type LawyerRepository interface {
	Create(lawyer *models.Lawyer) error
	FindByID(id uuid.UUID) (*models.Lawyer, error)
	FindByUserID(userID uuid.UUID) (*models.Lawyer, error)
	FindAll(filter LawyerFilter) ([]*models.Lawyer, int64, error)
	Update(lawyer *models.Lawyer) error
	UpdateAvailability(id uuid.UUID, available bool) error
}

type lawyerRepository struct{ db *sql.DB }

func NewLawyerRepository(db *sql.DB) LawyerRepository {
	return &lawyerRepository{db: db}
}

func (r *lawyerRepository) Create(l *models.Lawyer) error {
	l.ID = uuid.New()
	now := time.Now()
	l.CreatedAt = now
	l.UpdatedAt = now
	_, err := r.db.Exec(`
		INSERT INTO lawyers (id, user_id, license_number, specialization, years_of_experience,
		education, bio, office_address, city, province, latitude, longitude,
		consultation_fee_per_hour, is_available, is_verified, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		l.ID, l.UserID, l.LicenseNumber, pq.Array(l.Specialization),
		l.YearsOfExperience, l.Education, l.Bio, l.OfficeAddress, l.City, l.Province,
		l.Latitude, l.Longitude, l.ConsultationFeePerHour, l.IsAvailable, l.IsVerified,
		l.CreatedAt, l.UpdatedAt,
	)
	return err
}

func (r *lawyerRepository) FindByID(id uuid.UUID) (*models.Lawyer, error) {
	return r.scanOne(`SELECT * FROM lawyers WHERE id=$1`, id)
}

func (r *lawyerRepository) FindByUserID(userID uuid.UUID) (*models.Lawyer, error) {
	return r.scanOne(`SELECT * FROM lawyers WHERE user_id=$1`, userID)
}

func (r *lawyerRepository) FindAll(f LawyerFilter) ([]*models.Lawyer, int64, error) {
	args := []interface{}{}
	conditions := []string{"1=1"}
	idx := 1

	if f.Specialization != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(specialization)", idx))
		args = append(args, f.Specialization)
		idx++
	}
	if f.City != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(city) LIKE $%d", idx))
		args = append(args, "%"+strings.ToLower(f.City)+"%")
		idx++
	}
	if f.MinRating > 0 {
		conditions = append(conditions, fmt.Sprintf("rating >= $%d", idx))
		args = append(args, f.MinRating)
		idx++
	}
	if f.MaxFee > 0 {
		conditions = append(conditions, fmt.Sprintf("consultation_fee_per_hour <= $%d", idx))
		args = append(args, f.MaxFee)
		idx++
	}
	if f.MinFee > 0 {
		conditions = append(conditions, fmt.Sprintf("consultation_fee_per_hour >= $%d", idx))
		args = append(args, f.MinFee)
		idx++
	}
	if f.Search != "" {
		// joined with users via subquery
		conditions = append(conditions, fmt.Sprintf(`user_id IN (SELECT id FROM users WHERE LOWER(full_name) LIKE $%d)`, idx))
		args = append(args, "%"+strings.ToLower(f.Search)+"%")
		idx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int64
	if err := r.db.QueryRow("SELECT COUNT(*) FROM lawyers "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := fmt.Sprintf("SELECT * FROM lawyers %s ORDER BY rating DESC LIMIT $%d OFFSET $%d", where, idx, idx+1)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lawyers []*models.Lawyer
	for rows.Next() {
		l := &models.Lawyer{}
		if err := r.scanRow(rows, l); err != nil {
			return nil, 0, err
		}
		lawyers = append(lawyers, l)
	}
	return lawyers, total, nil
}

func (r *lawyerRepository) Update(l *models.Lawyer) error {
	_, err := r.db.Exec(`
		UPDATE lawyers SET specialization=$1, years_of_experience=$2, education=$3, bio=$4,
		office_address=$5, city=$6, province=$7, latitude=$8, longitude=$9,
		consultation_fee_per_hour=$10, updated_at=$11 WHERE id=$12`,
		pq.Array(l.Specialization), l.YearsOfExperience, l.Education, l.Bio,
		l.OfficeAddress, l.City, l.Province, l.Latitude, l.Longitude,
		l.ConsultationFeePerHour, time.Now(), l.ID,
	)
	return err
}

func (r *lawyerRepository) UpdateAvailability(id uuid.UUID, available bool) error {
	_, err := r.db.Exec(`UPDATE lawyers SET is_available=$1, updated_at=$2 WHERE id=$3`, available, time.Now(), id)
	return err
}

func (r *lawyerRepository) scanOne(query string, args ...interface{}) (*models.Lawyer, error) {
	row := r.db.QueryRow(query, args...)
	l := &models.Lawyer{}
	err := r.scanRow(row, l)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return l, err
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *lawyerRepository) scanRow(s scannable, l *models.Lawyer) error {
	return s.Scan(
		&l.ID, &l.UserID, &l.LicenseNumber, pq.Array(&l.Specialization),
		&l.YearsOfExperience, &l.Education, &l.Bio, &l.OfficeAddress,
		&l.City, &l.Province, &l.Latitude, &l.Longitude,
		&l.ConsultationFeePerHour, &l.IsAvailable, &l.Rating,
		&l.TotalReviews, &l.TotalConsultations, &l.IsVerified,
		&l.VerificationDocumentURL, &l.CreatedAt, &l.UpdatedAt,
	)
}
