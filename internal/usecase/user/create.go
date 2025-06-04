package user

import (
	"context"
	_ "cursach/internal/models"
	"cursach/internal/repository"
	"errors"
	"fmt"
)

var (
	ErrEmptyCredentials   = errors.New("login and password cannot be empty")
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrInvalidRole        = errors.New("invalid user role")
)

// UserManager определяет интерфейс для управления пользователями
type UserManager struct {
	userRepo repository.UserRepository
}

// NewUserManager создает новый экземпляр UserManager
func NewUserManager(userRepo repository.UserRepository) *UserManager {
	return &UserManager{userRepo: userRepo}
}

// CreateOrGetUser создает нового пользователя или возвращает существующего
// При создании проверяет уникальность логина
// При получении проверяет пароль
func (uc *UserManager) CreateOrGetUser(ctx context.Context, login, password, role string) (string, error) {
	// Валидация входных данных
	if login == "" || password == "" {
		return "", ErrEmptyCredentials
	}

	// Пытаемся получить существующего пользователя
	user, err := uc.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Если пользователь существует - проверяем пароль
	if user != nil {
		// Здесь должна быть реальная проверка хеша пароля
		if user.Password != password {
			return "", ErrInvalidCredentials
		}
		return user.ID, nil
	}

	// Если пользователя нет - создаем нового
	if role != "user" && role != "admin" {
		return "", ErrInvalidRole
	}

	userID, err := uc.userRepo.CreateUser(ctx, login, password, role)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}
