package messaging

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ParticipantIDs []primitive.ObjectID `bson:"participantIds" json:"participantIds"`
	CreatedAt      time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt" json:"updatedAt"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversationId" json:"conversationId"`
	SenderID       primitive.ObjectID `bson:"senderId" json:"senderId"`
	Content        string             `bson:"content" json:"content"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
}

type CreateConversationRequest struct {
	ParticipantIDs []primitive.ObjectID `json:"participantIds"`
}

type SendMessageRequest struct {
	ConversationID primitive.ObjectID `json:"conversationId"`
	Content        string             `json:"content"`
}

type MessageFrame struct {
	Type string `json:"type"` // "message"|"typing"|"read"
	// message
	ConversationID string `json:"conversationId,omitempty"`
	Content        string `json:"content,omitempty"`
}
