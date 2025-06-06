package message

import (
	"context"
	"cursach/internal/models"
	"cursach/internal/repository"
	"errors"
)

var (
	ErrEmptyMessage  = errors.New("message text cannot be empty")
	ErrUserNotInChat = errors.New("user not in chat")
)

type Sender struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
}

func NewSender(chatRepo repository.ChatRepository, messageRepo repository.MessageRepository) *Sender {
	return &Sender{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

func (uc *Sender) Execute(ctx context.Context, chatID, userID, text string) (*models.Message, error) {
	if text == "" {
		return nil, ErrEmptyMessage
	}

	isMember, err := uc.chatRepo.IsUserInChat(ctx, chatID, userID)
	if err != nil || !isMember {
		return nil, ErrUserNotInChat
	}

	msg := &models.Message{
		ChatID: chatID,
		UserID: userID,
		Text:   text,
	}

	// Исправлено: добавлено получение ID созданного сообщения
	messageID, err := uc.messageRepo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Устанавливаем ID сообщения
	msg.ID = messageID

	// Возвращаем полную модель сообщения
	return msg, nil
}
