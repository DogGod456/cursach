package user

import (
	"context"
	"cursach/internal/models"
)

func (m *UserManager) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	return m.userRepo.GetUserByID(ctx, userID)
}
