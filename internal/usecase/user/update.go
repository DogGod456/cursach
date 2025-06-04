package user

import (
	"context"
	"errors"

	"cursach/internal/repository"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
)

type LoginUpdater struct {
	userRepo repository.UserRepository
}

func NewLoginUpdater(userRepo repository.UserRepository) *LoginUpdater {
	return &LoginUpdater{userRepo: userRepo}
}

func (uc *LoginUpdater) UpdateLogin(ctx context.Context, userID, newLogin string) error {
	// Проверяем существование нового логина
	exists, err := uc.userRepo.LoginExists(ctx, newLogin)
	if err != nil {
		return err
	}
	if exists {
		return ErrLoginAlreadyExists
	}

	// Обновляем логин
	return uc.userRepo.UpdateLogin(ctx, userID, newLogin)
}
