package usecases

import (
	"context"
	"errors"
	"strings"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessagingUsecase struct {
	repo     msgpkg.IMessagingRepository
	userRepo userpkg.IUserRepository
}

func NewMessagingUsecase(repo msgpkg.IMessagingRepository, userRepo userpkg.IUserRepository) *MessagingUsecase {
	return &MessagingUsecase{repo: repo, userRepo: userRepo}
}

func (uc *MessagingUsecase) CreateConversation(ctx context.Context, userID primitive.ObjectID, participantIDs []primitive.ObjectID) (msgpkg.Conversation, error) {
	// ensure current user included
	found := false
	for _, id := range participantIDs {
		if id == userID {
			found = true
			break
		}
	}
	if !found {
		participantIDs = append(participantIDs, userID)
	}
	return uc.repo.CreateConversation(ctx, participantIDs)
}

func (uc *MessagingUsecase) GetUserConversations(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]msgpkg.Conversation, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.GetUserConversations(ctx, userID, limit, offset)
}

func (uc *MessagingUsecase) SendMessage(ctx context.Context, senderID primitive.ObjectID, conversationID primitive.ObjectID, content string) (msgpkg.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return msgpkg.Message{}, errors.New("message content cannot be empty")
	}
	// verify sender is participant
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return msgpkg.Message{}, err
	}
	isMember := false
	for _, id := range conv.ParticipantIDs {
		if id == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return msgpkg.Message{}, errors.New("forbidden")
	}
	return uc.repo.SaveMessage(ctx, msgpkg.Message{ConversationID: conversationID, SenderID: senderID, Content: content})
}

func (uc *MessagingUsecase) GetMessages(ctx context.Context, userID, conversationID primitive.ObjectID, limit, offset int) ([]msgpkg.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	isMember := false
	for _, id := range conv.ParticipantIDs {
		if id == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("forbidden")
	}
	return uc.repo.GetMessages(ctx, conversationID, limit, offset)
}

func (uc *MessagingUsecase) GetConversation(ctx context.Context, id primitive.ObjectID) (msgpkg.Conversation, error) {
	return uc.repo.GetConversation(ctx, id)
}

var _ msgpkg.IMessagingUsecase = (*MessagingUsecase)(nil)
