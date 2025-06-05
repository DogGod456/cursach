package user

import (
	"context"
	"cursach/internal/repository"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrForbidden    = errors.New("operation not permitted")
)

// UserDeleter определяет интерфейс для удаления пользователей
type UserDeleter struct {
	userRepo repository.UserRepository
}

// NewUserDeleter создает новый экземпляр UserDeleter
func NewUserDeleter(userRepo repository.UserRepository) *UserDeleter {
	return &UserDeleter{userRepo: userRepo}
}

// Execute удаляет пользователя по ID (без проверки прав)
func (uc *UserDeleter) Execute(ctx context.Context, userID string) error {
	// Проверка существования пользователя
	exists, err := uc.userRepo.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	// Удаление пользователя
	if err := uc.userRepo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
