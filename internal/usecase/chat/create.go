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
func (uc *ChatCreator) Execute(ctx context.Context, currentUserID, targetUserLogin string) (string, error) {
	targetUser, err := uc.userRepo.GetUserByLogin(ctx, targetUserLogin)
	if err != nil {
		return "", fmt.Errorf("failed to get user by login: %w", err)
	}
	if targetUser == nil {
		return "", ErrUserNotFound
	}

	// Проверяем, что пользователи разные
	if currentUserID == targetUser.ID {
		return "", errors.New("cannot create chat with yourself")
	}

	// Создаем чат с двумя участниками (без проверки существующих чатов)
	chatID, err := uc.chatRepo.CreateChat(ctx, []string{currentUserID, targetUser.ID})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrChatCreation, err)
	}

	return chatID, nil
}
