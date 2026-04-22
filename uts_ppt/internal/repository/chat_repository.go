package repository

import (
	"database/sql"
	"time"

	"legal-consultation-api/internal/models"

	"github.com/google/uuid"
)

type ChatRepository interface {
	SendMessage(msg *models.ChatMessage) error
	GetMessages(consultationID uuid.UUID, page, limit int) ([]*models.ChatMessage, int64, error)
	MarkRead(consultationID uuid.UUID, readerID uuid.UUID) error
}

type chatRepository struct{ db *sql.DB }

func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) SendMessage(msg *models.ChatMessage) error {
	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	_, err := r.db.Exec(`INSERT INTO chat_messages
		(id, consultation_id, sender_id, message_type, content, file_url, file_name, is_read, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		msg.ID, msg.ConsultationID, msg.SenderID, msg.MessageType,
		msg.Content, msg.FileURL, msg.FileName, false, msg.CreatedAt)
	return err
}

func (r *chatRepository) GetMessages(consultationID uuid.UUID, page, limit int) ([]*models.ChatMessage, int64, error) {
	var total int64
	r.db.QueryRow(`SELECT COUNT(*) FROM chat_messages WHERE consultation_id=$1`, consultationID).Scan(&total)
	offset := (page - 1) * limit
	rows, err := r.db.Query(`SELECT id, consultation_id, sender_id, message_type, content,
		file_url, file_name, is_read, read_at, created_at
		FROM chat_messages WHERE consultation_id=$1
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`, consultationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var msgs []*models.ChatMessage
	for rows.Next() {
		m := &models.ChatMessage{}
		rows.Scan(&m.ID, &m.ConsultationID, &m.SenderID, &m.MessageType,
			&m.Content, &m.FileURL, &m.FileName, &m.IsRead, &m.ReadAt, &m.CreatedAt)
		msgs = append(msgs, m)
	}
	return msgs, total, nil
}

func (r *chatRepository) MarkRead(consultationID uuid.UUID, readerID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE chat_messages SET is_read=true, read_at=$1
		WHERE consultation_id=$2 AND sender_id != $3 AND is_read=false`,
		now, consultationID, readerID)
	return err
}
