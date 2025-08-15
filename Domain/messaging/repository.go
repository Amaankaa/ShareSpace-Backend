package messaging

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockery --name=IMessagingRepository --output=../../mocks --outpkg=mocks

type IMessagingRepository interface {
	CreateConversation(ctx context.Context, participantIDs []primitive.ObjectID) (Conversation, error)
	GetUserConversations(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]Conversation, error)
	GetConversation(ctx context.Context, id primitive.ObjectID) (Conversation, error)

	SaveMessage(ctx context.Context, msg Message) (Message, error)
	GetMessages(ctx context.Context, conversationID primitive.ObjectID, limit, offset int) ([]Message, error)
}
