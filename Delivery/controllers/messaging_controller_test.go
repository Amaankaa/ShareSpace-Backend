package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	msgpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/messaging"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMessagingController_CreateConversation_AuthAndBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.IMessagingUsecase)
	ctrl := NewMessagingController(mockUC)

	// Unauthorized
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/conversations", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.CreateConversation(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// Bad request (invalid JSON)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("user_id", primitive.NewObjectID().Hex())
	c.Request = httptest.NewRequest(http.MethodPost, "/conversations", strings.NewReader(`{`))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.CreateConversation(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMessagingController_GetConversations_Auth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.IMessagingUsecase)
	ctrl := NewMessagingController(mockUC)

	// Unauthorized
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations", nil)
	ctrl.GetConversations(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMessagingController_GetMessages_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.IMessagingUsecase)
	ctrl := NewMessagingController(mockUC)

	// Unauthorized
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/abc/messages", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	ctrl.GetMessages(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	// Bad conversation ID
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("user_id", primitive.NewObjectID().Hex())
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/abc/messages", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	ctrl.GetMessages(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMessagingController_SuccessPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.IMessagingUsecase)
	ctrl := NewMessagingController(mockUC)

	// CreateConversation success
	uid := primitive.NewObjectID()
	p1 := primitive.NewObjectID().Hex()
	p2 := primitive.NewObjectID().Hex()
	ids := []primitive.ObjectID{}
	for _, s := range []string{p1, p2} {
		id, _ := primitive.ObjectIDFromHex(s)
		ids = append(ids, id)
	}
	conv := msgpkg.Conversation{ID: primitive.NewObjectID()}
	mockUC.On("CreateConversation", mock.Anything, uid, ids).Return(conv, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uid.Hex())
	body := `{"participantIds":["` + p1 + `","` + p2 + `"]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/conversations", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.CreateConversation(c)
	require.Equal(t, http.StatusCreated, w.Code)

	// GetConversations success
	mockUC.ExpectedCalls = nil
	mockUC.Calls = nil
	convID := primitive.NewObjectID()
	mockUC.On("GetUserConversations", mock.Anything, uid, 20, 0).Return([]msgpkg.Conversation{{ID: convID}}, nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("user_id", uid.Hex())
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations", nil)
	ctrl.GetConversations(c)
	require.Equal(t, http.StatusOK, w.Code)

	// GetMessages success
	mockUC.ExpectedCalls = nil
	mockUC.Calls = nil
	messages := []msgpkg.Message{{Content: "hi"}}
	mockUC.On("GetMessages", mock.Anything, uid, convID, 20, 0).Return(messages, nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("user_id", uid.Hex())
	c.Params = gin.Params{{Key: "id", Value: convID.Hex()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/"+convID.Hex()+"/messages", nil)
	ctrl.GetMessages(c)
	require.Equal(t, http.StatusOK, w.Code)
}
