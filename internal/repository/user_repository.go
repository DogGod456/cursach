package repository

import (
	"context"
	"cursach/internal/models"
	"database/sql"
	"errors"
	"fmt"
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
