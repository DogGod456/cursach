package user

import (
	"context"
	"cursach/internal/pkg/auth"
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
	salt     string
}

// NewUserManager создает новый экземпляр UserManager
func NewUserManager(userRepo repository.UserRepository, salt string) *UserManager {
	return &UserManager{
		userRepo: userRepo,
		salt:     salt,
	}
}

// CreateOrGetUser создает нового пользователя или возвращает существующего
// При создании:
//   - Проверяет уникальность логина
//   - Валидирует роль
//   - Хеширует пароль с использованием соли
//
// При получении:
//   - Проверяет пароль с помощью верификации хеша
//
// Возвращает идентификатор пользователя или ошибку
func (uc *UserManager) CreateOrGetUser(ctx context.Context, login, password, role string) (string, error) {
	// Валидация входных данных
	if login == "" || password == "" {
		return "", ErrEmptyCredentials
	}

	// Пытаемся получить существующего пользователя
	existingUser, err := uc.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Если пользователь существует - проверяем пароль
	if existingUser != nil {
		if !auth.VerifyPassword(password, uc.salt, existingUser.Password) {
			return "", ErrInvalidCredentials
		}
		return existingUser.ID, nil
	}

	// Если пользователя нет - создаем нового
	if role != "user" && role != "admin" {
		return "", ErrInvalidRole
	}

	// Хешируем пароль перед сохранением
	hashedPassword, err := auth.HashPassword(password, uc.salt)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := uc.userRepo.CreateUser(ctx, login, hashedPassword, role)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}
