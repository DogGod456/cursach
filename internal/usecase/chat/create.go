package chat

import (
	"context"
	"cursach/internal/repository"
	"errors"
	"fmt"
)

var (
	ErrEmptyUsers      = errors.New("at least one user required for chat creation")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserCheckFailed = errors.New("failed to verify user existence")
	ErrChatCreation    = errors.New("failed to create chat")
)

// ChatCreator определяет интерфейс для создания чатов
type ChatCreator struct {
	chatRepo repository.ChatRepository
	userRepo repository.UserRepository
}

// NewChatCreator создает новый экземпляр ChatCreator
func NewChatCreator(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
) *ChatCreator {
	return &ChatCreator{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

// Execute создает новый чат с указанными пользователями
// Возвращает ID созданного чата или ошибку
func (uc *ChatCreator) Execute(ctx context.Context, userIDs []string) (string, error) {
	// Проверка минимального количества пользователей
	if len(userIDs) == 0 {
		return "", ErrEmptyUsers
	}

	// Проверка существования всех пользователей
	for _, userID := range userIDs {
		exists, err := uc.userRepo.UserExists(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrUserCheckFailed, err)
		}
		if !exists {
			return "", fmt.Errorf("%w: user ID %s", ErrUserNotFound, userID)
		}
	}

	// Создание чата с пользователями
	chatID, err := uc.chatRepo.CreateChatWithUsers(ctx, userIDs...)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrChatCreation, err)
	}

	return chatID, nil
}
