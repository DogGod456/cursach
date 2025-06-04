package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
)

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByChat(ctx context.Context, chatID string, limit int) ([]models.Message, error)
}

type messageRepo struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepo{db: db}
}

func (r *messageRepo) Create(ctx context.Context, message *models.Message) error {
	query := `INSERT INTO messages (id_chat, id_user, message_text) 
              VALUES ($1, $2, $3) 
              RETURNING id_message, sending_time`
	return r.db.QueryRowContext(
		ctx,
		query,
		message.ChatID,
		message.UserID,
		message.Text,
	).Scan(&message.ID, &message.SendingTime)
}

func (r *messageRepo) GetByChat(ctx context.Context, chatID string, limit int) ([]models.Message, error) {
	query := `SELECT id_message, id_user, message_text, sending_time 
              FROM messages 
              WHERE id_chat = $1 
              ORDER BY sending_time DESC 
              LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Text, &m.SendingTime); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
