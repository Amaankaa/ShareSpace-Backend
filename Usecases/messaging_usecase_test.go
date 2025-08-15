package usecases_test

import (
	"context"
	"testing"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMessagingUsecase_SendMessage_MembershipAndContent(t *testing.T) {
	ctx := context.Background()
	repo := new(mocks.IMessagingRepository)
	userRepo := new(mocks.IUserRepository)
	uc := usecases.NewMessagingUsecase(repo, userRepo)

	convID := primitive.NewObjectID()
	senderID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()

	conv := msgpkg.Conversation{ID: convID, ParticipantIDs: []primitive.ObjectID{senderID, otherID}}
	repo.On("GetConversation", ctx, convID).Return(conv, nil)

	// empty content -> error
	_, err := uc.SendMessage(ctx, senderID, convID, "  ")
	require.Error(t, err)

	// valid content -> saved
	savedMsg := msgpkg.Message{ID: primitive.NewObjectID(), ConversationID: convID, SenderID: senderID, Content: "hi"}
	repo.On("SaveMessage", ctx, mock.MatchedBy(func(m msgpkg.Message) bool {
		return m.ConversationID == convID && m.SenderID == senderID && m.Content == "hi"
	})).Return(savedMsg, nil)
	msg, err := uc.SendMessage(ctx, senderID, convID, "hi")
	require.NoError(t, err)
	require.Equal(t, "hi", msg.Content)

	// non-member -> forbidden
	stranger := primitive.NewObjectID()
	repo.On("GetConversation", ctx, convID).Return(conv, nil)
	_, err = uc.SendMessage(ctx, stranger, convID, "hey")
	require.Error(t, err)
}

func TestMessagingUsecase_GetMessages_Membership(t *testing.T) {
	ctx := context.Background()
	repo := new(mocks.IMessagingRepository)
	userRepo := new(mocks.IUserRepository)
	uc := usecases.NewMessagingUsecase(repo, userRepo)

	convID := primitive.NewObjectID()
	member := primitive.NewObjectID()
	conv := msgpkg.Conversation{ID: convID, ParticipantIDs: []primitive.ObjectID{member}}
	repo.On("GetConversation", ctx, convID).Return(conv, nil)

	repo.On("GetMessages", ctx, convID, 20, 0).Return([]msgpkg.Message{}, nil)
	_, err := uc.GetMessages(ctx, member, convID, 20, 0)
	require.NoError(t, err)

	repo.On("GetConversation", ctx, convID).Return(conv, nil)
	_, err = uc.GetMessages(ctx, primitive.NewObjectID(), convID, 20, 0)
	require.Error(t, err)
}

// no custom matcher helpers needed
