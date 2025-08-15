package messaging

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockery --name=IMessagingUsecase --output=../../mocks --outpkg=mocks

type IMessagingUsecase interface {
	CreateConversation(ctx context.Context, userID primitive.ObjectID, participantIDs []primitive.ObjectID) (Conversation, error)
	GetUserConversations(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Conversation, error)
	GetConversation(ctx context.Context, id primitive.ObjectID) (Conversation, error)

	SendMessage(ctx context.Context, senderID primitive.ObjectID, conversationID primitive.ObjectID, content string) (Message, error)
	GetMessages(ctx context.Context, userID, conversationID primitive.ObjectID, limit, offset int) ([]Message, error)
}
