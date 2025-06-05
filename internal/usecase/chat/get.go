package chat

import (
	"context"
	"cursach/internal/models"
	"cursach/internal/repository"
)

type ChatLister struct {
	chatRepo repository.ChatRepository
}

func NewChatLister(chatRepo repository.ChatRepository) *ChatLister {
	return &ChatLister{chatRepo: chatRepo}
}

func (uc *ChatLister) Execute(ctx context.Context, userID string) ([]*models.ChatWithUser, error) {
	chats, err := uc.chatRepo.GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}
	return chats, nil
}
