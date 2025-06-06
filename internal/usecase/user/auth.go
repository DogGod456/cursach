package user

import (
	"context"
	"cursach/internal/models"
	"cursach/internal/pkg/auth"
	"cursach/internal/repository"
)

// Authenticator отвечает за аутентификацию пользователей
// Проверяет соответствие предоставленных учетных данных данным в системе
type Authenticator struct {
	userRepo repository.UserRepository
	salt     string
}

// NewAuthenticator создает новый экземпляр аутентификатора
// Возвращает инициализированный объект Authenticator
func NewAuthenticator(userRepo repository.UserRepository, salt string) *Authenticator {
	return &Authenticator{
		userRepo: userRepo,
		salt:     salt,
	}
}

// Authenticate выполняет аутентификацию пользователя по логину и паролю
func (uc *Authenticator) Authenticate(ctx context.Context, login, password string) (*models.User, error) {
	user, err := uc.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !auth.VerifyPassword(password, uc.salt, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
