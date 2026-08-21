package usecase

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/abijith/moneymate-backend/services/support/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ChatUseCase interface {
	SendMessage(ctx context.Context, senderID uuid.UUID, senderType string, receiverID uuid.UUID, receiverType string, message string) (*domain.ChatMessage, error)
	GetChatHistory(ctx context.Context, userID, adminID uuid.UUID) ([]*domain.ChatMessage, error)
	GetAdminChatHistory(ctx context.Context, adminID uuid.UUID) ([]*domain.ChatMessage, error)
	MarkMessagesAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error
}

type chatUseCase struct {
	repo domain.ChatRepository
	rdb  *redis.Client
}

func NewChatUseCase(repo domain.ChatRepository, rdb *redis.Client) ChatUseCase {
	return &chatUseCase{
		repo: repo,
		rdb:  rdb,
	}
}

func (u *chatUseCase) SendMessage(ctx context.Context, senderID uuid.UUID, senderType string, receiverID uuid.UUID, receiverType string, message string) (*domain.ChatMessage, error) {
	// Create chat message in DB
	msg, err := u.repo.CreateChatMessage(ctx, &domain.ChatMessage{
		SenderID:     senderID,
		SenderType:   senderType,
		ReceiverID:   receiverID,
		ReceiverType: receiverType,
		Message:      message,
	})
	if err != nil {
		return nil, err
	}

	// Publish to Redis
	payload, _ := json.Marshal(msg)
	u.rdb.Publish(ctx, "chat_messages", payload)

	return msg, nil
}

func (u *chatUseCase) GetChatHistory(ctx context.Context, userID, adminID uuid.UUID) ([]*domain.ChatMessage, error) {
	return u.repo.GetChatHistory(ctx, userID, adminID)
}

func (u *chatUseCase) GetAdminChatHistory(ctx context.Context, adminID uuid.UUID) ([]*domain.ChatMessage, error) {
	return u.repo.GetAdminChatHistory(ctx, adminID)
}

func (u *chatUseCase) MarkMessagesAsRead(ctx context.Context, senderID, receiverID uuid.UUID) error {
	return u.repo.MarkMessagesAsRead(ctx, senderID, receiverID)
}
