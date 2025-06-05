package chat

import (
	"context"
	"cursach/internal/repository"
	"errors"
	"fmt"
	"log"
)

var (
	ErrChatNotFound  = errors.New("chat not found")
	ErrUserNotInChat = errors.New("user is not a participant of the chat")
	ErrChatDeletion  = errors.New("failed to delete chat")
)

// ChatDeleter определяет интерфейс для удаления чатов
type ChatDeleter struct {
	chatRepo repository.ChatRepository
}

// NewChatDeleter создает новый экземпляр ChatDeleter
func NewChatDeleter(chatRepo repository.ChatRepository) *ChatDeleter {
	return &ChatDeleter{chatRepo: chatRepo}
}

// Execute удаляет чат, если пользователь является его участником
func (uc *ChatDeleter) Execute(ctx context.Context, chatID, userID string) error {
	// Проверяем, что пользователь является участником чата
	isMember, err := uc.chatRepo.IsUserInChat(ctx, chatID, userID)
	log.Println(isMember, err)
	if err != nil {
		return fmt.Errorf("failed to check user membership: %w", err)
	}
	if !isMember {
		return ErrUserNotInChat
	}

	// Удаляем чат
	if err := uc.chatRepo.DeleteChat(ctx, chatID); err != nil {
		return fmt.Errorf("%w: %v", ErrChatDeletion, err)
	}

	return nil
}
