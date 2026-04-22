package repository

import (
	"database/sql"
	"errors"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	Update(user *models.User) error
	UpdateLastLogin(id uuid.UUID) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, full_name, phone, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	user.ID = uuid.New()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.db.Exec(query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.FullName,
		user.Phone, user.IsActive, user.IsVerified, user.CreatedAt, user.UpdatedAt,
	)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return errors.New("email already registered")
	}
	return err
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password_hash, role, full_name, phone, profile_photo_url,
		is_active, is_verified, last_login_at, created_at, updated_at FROM users WHERE email=$1`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.FullName,
		&user.Phone, &user.ProfilePhotoURL, &user.IsActive, &user.IsVerified,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password_hash, role, full_name, phone, profile_photo_url,
		is_active, is_verified, last_login_at, created_at, updated_at FROM users WHERE id=$1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.FullName,
		&user.Phone, &user.ProfilePhotoURL, &user.IsActive, &user.IsVerified,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *userRepository) Update(user *models.User) error {
	query := `UPDATE users SET full_name=$1, phone=$2, profile_photo_url=$3, updated_at=$4 WHERE id=$5`
	_, err := r.db.Exec(query, user.FullName, user.Phone, user.ProfilePhotoURL, time.Now(), user.ID)
	return err
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID) error {
	_, err := r.db.Exec(`UPDATE users SET last_login_at=$1 WHERE id=$2`, time.Now(), id)
	return err
}
