package user

import (
	"context"

	"cursach/internal/repository"
)

type Logouter struct {
	tokenRepo repository.TokenRepository
}

func NewLogouter(tokenRepo repository.TokenRepository) *Logouter {
	return &Logouter{tokenRepo: tokenRepo}
}

func (uc *Logouter) Logout(ctx context.Context, token string) error {
	return uc.tokenRepo.RevokeToken(ctx, token)
}
