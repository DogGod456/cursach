package user

import (
	"context"
	"cursach/internal/models"
	"cursach/internal/repository"
)

type UserSearcher struct {
	userRepo repository.UserRepository
}

func NewUserSearcher(userRepo repository.UserRepository) *UserSearcher {
	return &UserSearcher{userRepo: userRepo}
}

func (uc *UserSearcher) Execute(ctx context.Context, login string) ([]*models.User, error) {
	users, err := uc.userRepo.SearchUsersByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	return users, nil
}
