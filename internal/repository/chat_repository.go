package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
	"errors"
	"fmt"
)

// ChatRepository определяет интерфейс для работы с чатами
type ChatRepository interface {
	// CreateChat создает новый чат и возвращает его ID
	CreateChat(ctx context.Context) (string, error)

	// DeleteChat удаляет чат по указанному ID
	DeleteChat(ctx context.Context, chatID string) error

	// GetChatByID возвращает чат по его ID
	GetChatByID(ctx context.Context, chatID string) (*models.Chat, error)

	// FindChatByUsers ищет чат, в котором состоят все указанные пользователи
	// Возвращает ID чата, если такой чат существует
	FindChatByUsers(ctx context.Context, userIDs ...string) (string, error)

	// GetChatUsers возвращает список пользователей, участвующих в указанном чате
	GetChatUsers(ctx context.Context, chatID string) ([]models.User, error)

	// IsUserInChat проверяет, состоит ли пользователь в указанном чате
	IsUserInChat(ctx context.Context, chatID, userID string) (bool, error)

	// CreateChatWithUsers создает новый чат и добавляет в него указанных пользователей
	CreateChatWithUsers(ctx context.Context, userIDs ...string) (string, error)
}

// chatRepository реализует интерфейс ChatRepository
type chatRepository struct {
	db *sql.DB
}

// NewChatRepository создает новый экземпляр ChatRepository
func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(ctx context.Context) (string, error) {
	var chatID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id_chat`,
	).Scan(&chatID)
	return chatID, err
}

func (r *chatRepository) DeleteChat(ctx context.Context, chatID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM chats WHERE id_chat = $1`,
		chatID,
	)
	return err
}

func (r *chatRepository) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.QueryRowContext(ctx,
		`SELECT id_chat, created_at, updated_at 
        FROM chats 
        WHERE id_chat = $1`,
		chatID,
	).Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) FindChatByUsers(ctx context.Context, userIDs ...string) (string, error) {
	if len(userIDs) == 0 {
		return "", errors.New("at least one user is required")
	}

	// Создаем параметры для запроса
	params := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		params[i] = id
	}

	// Генерируем плейсхолдеры для IN: $1, $2, ...
	inClause := ""
	for i := range userIDs {
		if i > 0 {
			inClause += ", "
		}
		inClause += fmt.Sprintf("$%d", i+1)
	}

	// Запрос: находим чаты, в которых есть все указанные пользователи
	query := fmt.Sprintf(`
        SELECT id_chat
        FROM chat_users
        WHERE id_user IN (%s)
        GROUP BY id_chat
        HAVING COUNT(DISTINCT id_user) = %d
    `, inClause, len(userIDs))

	var chatID string
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&chatID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return chatID, err
}

func (r *chatRepository) GetChatUsers(ctx context.Context, chatID string) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id_user, u.login, u.role, u.created_at, u.updated_at
        FROM users u
        JOIN chat_users cu ON u.id_user = cu.id_user
        WHERE cu.id_chat = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID,
			&u.Login,
			&u.Role,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *chatRepository) IsUserInChat(ctx context.Context, chatID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM chat_users 
            WHERE id_chat = $1 AND id_user = $2
        )`,
		chatID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *chatRepository) CreateChatWithUsers(ctx context.Context, userIDs ...string) (string, error) {
	if len(userIDs) < 1 {
		return "", errors.New("at least one user is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var chatID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id_chat`,
	).Scan(&chatID)
	if err != nil {
		return "", err
	}

	for _, userID := range userIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO chat_users (id_chat, id_user)
            VALUES ($1, $2)`,
			chatID, userID,
		)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return chatID, nil
}
