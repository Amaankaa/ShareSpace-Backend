package repositories

import (
	"context"
	"fmt"
	"time"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessagingRepository struct {
	convs *mongo.Collection
	msgs  *mongo.Collection
}

func NewMessagingRepository(conversations, messages *mongo.Collection) *MessagingRepository {
	return &MessagingRepository{convs: conversations, msgs: messages}
}

func (r *MessagingRepository) CreateConversation(ctx context.Context, participantIDs []primitive.ObjectID) (msgpkg.Conversation, error) {
	c := msgpkg.Conversation{ID: primitive.NewObjectID(), ParticipantIDs: participantIDs, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if _, err := r.convs.InsertOne(ctx, c); err != nil {
		return msgpkg.Conversation{}, fmt.Errorf("failed to create conversation: %w", err)
	}
	return c, nil
}

func (r *MessagingRepository) GetUserConversations(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]msgpkg.Conversation, error) {
	filter := bson.M{"participantIds": userID}
	opts := options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	cur, err := r.convs.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer cur.Close(ctx)
	var list []msgpkg.Conversation
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode conversations: %w", err)
	}
	return list, nil
}

func (r *MessagingRepository) GetConversation(ctx context.Context, id primitive.ObjectID) (msgpkg.Conversation, error) {
	var c msgpkg.Conversation
	if err := r.convs.FindOne(ctx, bson.M{"_id": id}).Decode(&c); err != nil {
		if err == mongo.ErrNoDocuments {
			return msgpkg.Conversation{}, fmt.Errorf("not found")
		}
		return msgpkg.Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}
	return c, nil
}

func (r *MessagingRepository) SaveMessage(ctx context.Context, msg msgpkg.Message) (msgpkg.Message, error) {
	msg.ID = primitive.NewObjectID()
	msg.CreatedAt = time.Now()
	if _, err := r.msgs.InsertOne(ctx, msg); err != nil {
		return msgpkg.Message{}, fmt.Errorf("failed to save message: %w", err)
	}
	// bump conversation updatedAt
	_, _ = r.convs.UpdateByID(ctx, msg.ConversationID, bson.M{"$set": bson.M{"updatedAt": time.Now()}})
	return msg, nil
}

func (r *MessagingRepository) GetMessages(ctx context.Context, conversationID primitive.ObjectID, limit, offset int) ([]msgpkg.Message, error) {
	filter := bson.M{"conversationId": conversationID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	cur, err := r.msgs.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer cur.Close(ctx)
	var list []msgpkg.Message
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}
	return list, nil
}
