package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
	"fmt"
)

// MessageRepository определяет интерфейс для работы с сообщениями
type MessageRepository interface {
	// Create создает новое сообщение и возвращает его ID
	Create(ctx context.Context, message *models.Message) (string, error)

	// GetByID возвращает сообщение по его ID
	GetByID(ctx context.Context, messageID string) (*models.Message, error)

	// GetByChat возвращает сообщения для указанного чата
	GetByChat(ctx context.Context, chatID string, limit int) ([]*models.Message, error)

	// Update обновляет текст сообщения - не используется, возможно в дальнейшем при развитии проекта
	Update(ctx context.Context, messageID, newText string) error

	// Delete удаляет сообщение по ID - не используется, возможно в дальнейшем при развитии проекта
	Delete(ctx context.Context, messageID string) error
}

// messageRepository реализует интерфейс MessageRepository
type messageRepository struct {
	db *sql.DB
}

// NewMessageRepository создает новый экземпляр MessageRepository
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) (string, error) {
	var messageID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO messages (id_chat, id_user, message_text) 
		VALUES ($1, $2, $3) 
		RETURNING id_message`,
		message.ChatID,
		message.UserID,
		message.Text,
	).Scan(&messageID)

	if err != nil {
		return "", fmt.Errorf("failed to create message: %w", err)
	}
	return messageID, nil
}

func (r *messageRepository) GetByID(ctx context.Context, messageID string) (*models.Message, error) {
	var msg models.Message
	err := r.db.QueryRowContext(ctx,
		`SELECT id_message, id_chat, id_user, message_text, sending_time, updated_at
		FROM messages
		WHERE id_message = $1`,
		messageID,
	).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.UserID,
		&msg.Text,
		&msg.SendingTime,
		&msg.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}
	return &msg, nil
}

func (r *messageRepository) GetByChat(ctx context.Context, chatID string, limit int) ([]*models.Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
			m.id_message, 
			m.id_chat, 
			m.id_user, 
			u.login,  -- Добавлен логин пользователя
			m.message_text, 
			m.sending_time, 
			m.updated_at
		FROM messages m
		JOIN users u ON m.id_user = u.id_user
		WHERE m.id_chat = $1
		ORDER BY m.sending_time DESC
		LIMIT $2`,
		chatID,
		limit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get messages by chat: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.UserID,
			&msg.Login, // Сканируем логин
			&msg.Text,
			&msg.SendingTime,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return messages, nil
}

func (r *messageRepository) Update(ctx context.Context, messageID, newText string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages 
		SET message_text = $1, updated_at = NOW()
		WHERE id_message = $2`,
		newText,
		messageID,
	)

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, messageID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM messages WHERE id_message = $1`,
		messageID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}
