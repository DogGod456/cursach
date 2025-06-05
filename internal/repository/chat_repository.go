package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// ChatRepository определяет интерфейс для работы с чатами
type ChatRepository interface {
	// CreateChat создает новый чат и возвращает его ID
	CreateChat(ctx context.Context, userIDs []string) (string, error)

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

	GetUserChats(ctx context.Context, userID string) ([]*models.ChatWithUser, error)
}

// chatRepository реализует интерфейс ChatRepository
type chatRepository struct {
	db *sql.DB
}

// NewChatRepository создает новый экземпляр ChatRepository
func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(ctx context.Context, userIDs []string) (string, error) {
	if len(userIDs) == 0 {
		return "", errors.New("at least one user is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Создаем чат
	var chatID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id_chat`,
	).Scan(&chatID)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	// Добавляем пользователей в чат
	for _, userID := range userIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO chat_users (id_chat, id_user) VALUES ($1, $2)`,
			chatID, userID,
		)
		if err != nil {
			return "", fmt.Errorf("failed to add user to chat: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return chatID, nil
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
	log.Println(chatID, userID, "is user in chat")
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

func (r *chatRepository) GetUserChats(ctx context.Context, userID string) ([]*models.ChatWithUser, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
            c.id_chat, 
            c.created_at, 
            c.updated_at,
            u.id_user,
            u.login
        FROM chats c
        JOIN chat_users uc ON c.id_chat = uc.id_chat
        JOIN users u ON uc.id_user = u.id_user
        WHERE uc.id_user != $1 AND c.id_chat IN (
            SELECT id_chat FROM chat_users WHERE id_user = $1
        )`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}
	defer rows.Close()

	var chats []*models.ChatWithUser
	for rows.Next() {
		var chat models.ChatWithUser
		if err := rows.Scan(
			&chat.Chat.ID,
			&chat.Chat.CreatedAt,
			&chat.Chat.UpdatedAt,
			&chat.User.ID,
			&chat.User.Login,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, &chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return chats, nil
}
