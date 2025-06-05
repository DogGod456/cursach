package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// UserRepository определяет интерфейс для работы с пользователями системы
type UserRepository interface {
	// CreateUser создает нового пользователя и возвращает его ID
	CreateUser(ctx context.Context, login, passwordHash, role string) (string, error)

	// GetUserByID возвращает пользователя по его ID
	GetUserByID(ctx context.Context, userID string) (*models.User, error)

	// GetUserByLogin возвращает пользователя по логину
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, user *models.User) error

	// DeleteUser удаляет пользователя по ID
	DeleteUser(ctx context.Context, userID string) error

	// UserExists проверяет существование пользователя по ID
	UserExists(ctx context.Context, userID string) (bool, error)

	// LoginExists проверяет существование пользователя по логину
	LoginExists(ctx context.Context, login string) (bool, error)

	// UpdateLogin обновляет логин пользователя
	UpdateLogin(ctx context.Context, userID, newLogin string) error

	SearchUsersByLogin(ctx context.Context, login string) ([]*models.User, error)
}

// userRepository реализует интерфейс UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, login, passwordHash, role string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (login, password_hash, role) 
		VALUES ($1, $2, $3) 
		RETURNING id_user`,
		login, passwordHash, role,
	).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id_user, login, password_hash, role, created_at, updated_at
		FROM users
		WHERE id_user = $1`,
		userID,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Запрос для получения личных чатов с собеседниками
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			c.id_chat, 
			c.created_at AS chat_created_at,
			c.updated_at AS chat_updated_at,
			u.id_user AS other_user_id, 
			u.login AS other_user_login, 
			u.role AS other_user_role, 
			u.created_at AS other_user_created_at, 
			u.updated_at AS other_user_updated_at
		FROM chats c
		JOIN chat_users cu1 ON c.id_chat = cu1.id_chat
		JOIN chat_users cu2 
			ON c.id_chat = cu2.id_chat 
			AND cu2.id_user != cu1.id_user
		JOIN users u ON u.id_user = cu2.id_user
		WHERE cu1.id_user = $1
			AND c.id_chat IN (
				SELECT id_chat 
				FROM chat_users 
				GROUP BY id_chat 
				HAVING COUNT(*) = 2
			)`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user chats: %w", err)
	}
	defer rows.Close()

	var chats []models.ChatWithUser
	for rows.Next() {
		var chat models.Chat
		var otherUser models.User
		var chatUpdatedAt, userUpdatedAt sql.NullTime

		err := rows.Scan(
			&chat.ID,
			&chat.CreatedAt,
			&chatUpdatedAt,
			&otherUser.ID,
			&otherUser.Login,
			&otherUser.Role,
			&otherUser.CreatedAt,
			&userUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat row: %w", err)
		}

		// Обработка NULL значений для времени обновления
		if chatUpdatedAt.Valid {
			chat.UpdatedAt = chatUpdatedAt
		}
		if userUpdatedAt.Valid {
			otherUser.UpdatedAt = userUpdatedAt
		}

		chats = append(chats, models.ChatWithUser{
			Chat: chat,
			User: otherUser,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	user.Chats = chats
	log.Println("чаты", chats)
	return &user, nil
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id_user, login, password_hash, role, created_at, updated_at
		FROM users
		WHERE login = $1`,
		login,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users 
		SET login = $1, password_hash = $2, role = $3, updated_at = NOW()
		WHERE id_user = $4`,
		user.Login,
		user.Password,
		user.Role,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM users WHERE id_user = $1`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id_user = $1)`,
		userID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

func (r *userRepository) LoginExists(ctx context.Context, login string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)`,
		login,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check login existence: %w", err)
	}
	return exists, nil
}

// UpdateLogin обновляет логин пользователя
func (r *userRepository) UpdateLogin(ctx context.Context, userID, newLogin string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users 
		SET login = $1, updated_at = NOW()
		WHERE id_user = $2`,
		newLogin,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user login: %w", err)
	}
	return nil
}

func (r *userRepository) SearchUsersByLogin(ctx context.Context, login string) ([]*models.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id_user, login, role, created_at, updated_at
		FROM users
		WHERE login LIKE $1 || '%'`,
		login,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search users by login: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
