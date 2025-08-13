package usecases_test

import (
	"context"
	"testing"
	"time"

	mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mentorshipUsecaseTestSuite struct {
	suite.Suite
	mockMentorshipRepo *mocks.IMentorshipRepository
	mockUserRepo       *mocks.IUserRepository
	usecase            *usecases.MentorshipUsecase
	ctx                context.Context
}

func TestMentorshipUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(mentorshipUsecaseTestSuite))
}

func (s *mentorshipUsecaseTestSuite) SetupTest() {
	s.mockMentorshipRepo = new(mocks.IMentorshipRepository)
	s.mockUserRepo = new(mocks.IUserRepository)
	s.usecase = usecases.NewMentorshipUsecase(s.mockMentorshipRepo, s.mockUserRepo)
	s.ctx = context.Background()
}

func (s *mentorshipUsecaseTestSuite) TestSendMentorshipRequest_Success() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	topics := []string{"Technology Skills", "Career Guidance"}

	request := mentorshippkg.CreateMentorshipRequestDTO{
		MentorID: mentorID,
		Message:  "I would like to learn from you",
		Topics:   topics,
	}

	// Mock mentor profile (available for mentoring)
	mentorProfile := userpkg.PublicProfile{
		ID:                    mentorID,
		DisplayName:           "TechMentor",
		IsMentor:              true,
		AvailableForMentoring: true,
		MentorshipTopics:      topics,
	}

	menteeProfile := userpkg.PublicProfile{
		ID:          menteeID,
		DisplayName: "TechStudent",
		IsMentee:    true,
	}

	// Setup mocks for CanSendRequest validation
	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)
	s.mockMentorshipRepo.On("ExistsPendingRequest", s.ctx, menteeID.Hex(), mentorID.Hex()).Return(false, nil)
	s.mockMentorshipRepo.On("GetRequestsByMentee", s.ctx, menteeID.Hex(), 11, 0).Return([]mentorshippkg.MentorshipRequest{}, nil)
	s.mockMentorshipRepo.On("GetActiveConnectionsByUser", s.ctx, mentorID.Hex()).Return([]mentorshippkg.MentorshipConnection{}, nil)

	// Mock request creation
	createdRequest := mentorshippkg.MentorshipRequest{
		ID:        primitive.NewObjectID(),
		MenteeID:  menteeID,
		MentorID:  mentorID,
		Status:    mentorshippkg.StatusPending,
		Message:   request.Message,
		Topics:    request.Topics,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.mockMentorshipRepo.On("CreateRequest", s.ctx, mock.AnythingOfType("mentorshippkg.MentorshipRequest")).Return(createdRequest, nil)

	// Mock profile retrieval for response building
	s.mockUserRepo.On("GetPublicProfile", s.ctx, menteeID.Hex()).Return(menteeProfile, nil)
	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)

	result, err := s.usecase.SendMentorshipRequest(s.ctx, menteeID.Hex(), request)

	s.NoError(err)
	s.Equal(createdRequest.ID, result.ID)
	s.Equal(mentorshippkg.StatusPending, result.Status)
	s.Equal("TechStudent", result.MenteeInfo.DisplayName)
	s.Equal("TechMentor", result.MentorInfo.DisplayName)
	s.mockMentorshipRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestSendMentorshipRequest_CannotRequestSelf() {
	userID := primitive.NewObjectID()
	request := mentorshippkg.CreateMentorshipRequestDTO{
		MentorID: userID,
		Topics:   []string{"Technology Skills"},
	}

	_, err := s.usecase.SendMentorshipRequest(s.ctx, userID.Hex(), request)

	s.Error(err)
	s.Equal(mentorshippkg.ErrCannotRequestSelf, err)
}

func (s *mentorshipUsecaseTestSuite) TestSendMentorshipRequest_MentorNotAvailable() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()

	request := mentorshippkg.CreateMentorshipRequestDTO{
		MentorID: mentorID,
		Topics:   []string{"Technology Skills"},
	}

	// Mock mentor profile (not available for mentoring)
	mentorProfile := userpkg.PublicProfile{
		ID:                    mentorID,
		DisplayName:           "UnavailableMentor",
		IsMentor:              true,
		AvailableForMentoring: false,
	}

	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)

	_, err := s.usecase.SendMentorshipRequest(s.ctx, menteeID.Hex(), request)

	s.Error(err)
	s.Equal(mentorshippkg.ErrMentorNotAvailable, err)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestSendMentorshipRequest_RequestAlreadyExists() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()

	request := mentorshippkg.CreateMentorshipRequestDTO{
		MentorID: mentorID,
		Topics:   []string{"Technology Skills"},
	}

	mentorProfile := userpkg.PublicProfile{
		ID:                    mentorID,
		DisplayName:           "TechMentor",
		IsMentor:              true,
		AvailableForMentoring: true,
	}

	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)
	s.mockMentorshipRepo.On("ExistsPendingRequest", s.ctx, menteeID.Hex(), mentorID.Hex()).Return(true, nil)

	_, err := s.usecase.SendMentorshipRequest(s.ctx, menteeID.Hex(), request)

	s.Error(err)
	s.Equal(mentorshippkg.ErrRequestAlreadyExists, err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockMentorshipRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestRespondToRequest_Accept() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()

	request := mentorshippkg.MentorshipRequest{
		ID:       requestID,
		MenteeID: menteeID,
		MentorID: mentorID,
		Status:   mentorshippkg.StatusPending,
		Topics:   []string{"Technology Skills"},
	}

	response := mentorshippkg.RespondToRequestDTO{
		Accept: true,
	}

	menteeProfile := userpkg.PublicProfile{
		ID:          menteeID,
		DisplayName: "TechStudent",
	}

	mentorProfile := userpkg.PublicProfile{
		ID:          mentorID,
		DisplayName: "TechMentor",
	}

	// Mock request retrieval for validation (first call in ValidateRequestAccess)
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(request, nil).Once()

	// Mock request retrieval for main logic (second call in RespondToRequest)
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(request, nil).Once()

	// Mock connection creation
	s.mockMentorshipRepo.On("CreateConnection", s.ctx, mock.AnythingOfType("mentorshippkg.MentorshipConnection")).Return(mentorshippkg.MentorshipConnection{}, nil)

	// Mock status update
	s.mockMentorshipRepo.On("UpdateRequestStatus", s.ctx, requestID.Hex(), mentorshippkg.StatusAccepted).Return(nil)

	// Mock updated request retrieval (third call to get updated request)
	updatedRequest := request
	updatedRequest.Status = mentorshippkg.StatusAccepted
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(updatedRequest, nil).Once()

	// Mock profile retrieval for response building
	s.mockUserRepo.On("GetPublicProfile", s.ctx, menteeID.Hex()).Return(menteeProfile, nil)
	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)

	result, err := s.usecase.RespondToRequest(s.ctx, requestID.Hex(), mentorID.Hex(), response)

	s.NoError(err)
	s.Equal(mentorshippkg.StatusAccepted, result.Status)
	s.mockMentorshipRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestRespondToRequest_Reject() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()

	request := mentorshippkg.MentorshipRequest{
		ID:       requestID,
		MenteeID: menteeID,
		MentorID: mentorID,
		Status:   mentorshippkg.StatusPending,
		Topics:   []string{"Technology Skills"},
	}

	response := mentorshippkg.RespondToRequestDTO{
		Accept: false,
		Reason: "Not available at the moment",
	}

	menteeProfile := userpkg.PublicProfile{
		ID:          menteeID,
		DisplayName: "TechStudent",
	}

	mentorProfile := userpkg.PublicProfile{
		ID:          mentorID,
		DisplayName: "TechMentor",
	}

	// Mock request retrieval for validation (first call in ValidateRequestAccess)
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(request, nil).Once()

	// Mock request retrieval for main logic (second call in RespondToRequest)
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(request, nil).Once()

	// Mock status update
	s.mockMentorshipRepo.On("UpdateRequestStatus", s.ctx, requestID.Hex(), mentorshippkg.StatusRejected).Return(nil)

	// Mock updated request retrieval (third call to get updated request)
	updatedRequest := request
	updatedRequest.Status = mentorshippkg.StatusRejected
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(updatedRequest, nil).Once()

	// Mock profile retrieval for response building
	s.mockUserRepo.On("GetPublicProfile", s.ctx, menteeID.Hex()).Return(menteeProfile, nil)
	s.mockUserRepo.On("GetPublicProfile", s.ctx, mentorID.Hex()).Return(mentorProfile, nil)

	result, err := s.usecase.RespondToRequest(s.ctx, requestID.Hex(), mentorID.Hex(), response)

	s.NoError(err)
	s.Equal(mentorshippkg.StatusRejected, result.Status)
	s.mockMentorshipRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestRespondToRequest_UnauthorizedUser() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	unauthorizedUserID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()

	request := mentorshippkg.MentorshipRequest{
		ID:       requestID,
		MenteeID: menteeID,
		MentorID: mentorID,
		Status:   mentorshippkg.StatusPending,
		Topics:   []string{"Technology Skills"},
	}

	response := mentorshippkg.RespondToRequestDTO{
		Accept: true,
	}

	// Mock request retrieval for validation
	s.mockMentorshipRepo.On("GetRequestByID", s.ctx, requestID.Hex()).Return(request, nil)

	_, err := s.usecase.RespondToRequest(s.ctx, requestID.Hex(), unauthorizedUserID.Hex(), response)

	s.Error(err)
	s.Equal(mentorshippkg.ErrUnauthorizedAction, err)
	s.mockMentorshipRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestGetActiveConnections() {
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	connections := []mentorshippkg.MentorshipConnection{
		{
			ID:       primitive.NewObjectID(),
			MenteeID: userID,
			MentorID: otherUserID,
			Status:   mentorshippkg.ConnectionActive,
			Topics:   []string{"Technology Skills"},
		},
		{
			ID:       primitive.NewObjectID(),
			MenteeID: otherUserID,
			MentorID: userID,
			Status:   mentorshippkg.ConnectionActive,
			Topics:   []string{"Career Guidance"},
		},
	}

	userProfile := userpkg.PublicProfile{
		ID:          userID,
		DisplayName: "User",
	}

	otherUserProfile := userpkg.PublicProfile{
		ID:          otherUserID,
		DisplayName: "OtherUser",
	}

	s.mockMentorshipRepo.On("GetActiveConnectionsByUser", s.ctx, userID.Hex()).Return(connections, nil)
	s.mockUserRepo.On("GetPublicProfile", s.ctx, userID.Hex()).Return(userProfile, nil).Times(2)
	s.mockUserRepo.On("GetPublicProfile", s.ctx, otherUserID.Hex()).Return(otherUserProfile, nil).Times(2)

	result, err := s.usecase.GetActiveConnections(s.ctx, userID.Hex())

	s.NoError(err)
	s.Len(result, 2)
	s.Equal(mentorshippkg.ConnectionActive, result[0].Status)
	s.Equal(mentorshippkg.ConnectionActive, result[1].Status)
	s.mockMentorshipRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TestGetMentorshipStats() {
	userID := primitive.NewObjectID()

	expectedStats := mentorshippkg.MentorshipStats{
		UserID:                 userID,
		TotalMentorRequests:    5,
		AcceptedMentorRequests: 3,
		ActiveMentorships:      2,
		CompletedMentorships:   1,
		TotalMenteeRequests:    2,
		AcceptedMenteeRequests: 1,
		ActiveMenteerships:     1,
		CompletedMenteerships:  0,
		TotalConnections:       1,
	}

	s.mockMentorshipRepo.On("GetMentorshipStats", s.ctx, userID.Hex()).Return(expectedStats, nil)

	result, err := s.usecase.GetMentorshipStats(s.ctx, userID.Hex())

	s.NoError(err)
	s.Equal(expectedStats, result)
	s.mockMentorshipRepo.AssertExpectations(s.T())
}

func (s *mentorshipUsecaseTestSuite) TearDownTest() {
	s.mockMentorshipRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}
