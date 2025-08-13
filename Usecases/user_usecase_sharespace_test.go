package usecases_test

import (
	"context"
	"testing"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userShareSpaceUsecaseTestSuite struct {
	suite.Suite
	mockUserRepo          *mocks.IUserRepository
	mockPasswordSvc       *mocks.IPasswordService
	mockTokenRepo         *mocks.ITokenRepository
	mockJWTService        *mocks.IJWTService
	mockEmailVerifier     *mocks.IEmailVerifier
	mockEmailSender       *mocks.IEmailSender
	mockPasswordResetRepo *mocks.IPasswordResetRepository
	mockVerificationRepo  *mocks.IVerificationRepository
	mockCloudinaryService *mocks.ICloudinaryService
	usecase               *usecases.UserUsecase
	ctx                   context.Context
}

func TestUserShareSpaceUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(userShareSpaceUsecaseTestSuite))
}

func (s *userShareSpaceUsecaseTestSuite) SetupTest() {
	s.mockUserRepo = new(mocks.IUserRepository)
	s.mockPasswordSvc = new(mocks.IPasswordService)
	s.mockTokenRepo = new(mocks.ITokenRepository)
	s.mockJWTService = new(mocks.IJWTService)
	s.mockEmailVerifier = new(mocks.IEmailVerifier)
	s.mockEmailSender = new(mocks.IEmailSender)
	s.mockPasswordResetRepo = new(mocks.IPasswordResetRepository)
	s.mockVerificationRepo = new(mocks.IVerificationRepository)
	s.mockCloudinaryService = new(mocks.ICloudinaryService)

	s.usecase = usecases.NewUserUsecase(
		s.mockUserRepo,
		s.mockPasswordSvc,
		s.mockTokenRepo,
		s.mockJWTService,
		s.mockEmailVerifier,
		s.mockEmailSender,
		s.mockPasswordResetRepo,
		s.mockVerificationRepo,
		s.mockCloudinaryService,
	)

	s.ctx = context.Background()
}

func (s *userShareSpaceUsecaseTestSuite) TestGetPublicProfile_Success() {
	userID := primitive.NewObjectID().Hex()
	expectedProfile := userpkg.PublicProfile{
		ID:                    primitive.NewObjectID(),
		DisplayName:           "TechMentor",
		Bio:                   "I love helping with tech",
		IsMentor:              true,
		IsMentee:              false,
		MentorshipTopics:      []string{"Technology Skills"},
		MentorshipBio:         "Expert in programming",
		AvailableForMentoring: true,
	}

	s.mockUserRepo.On("GetPublicProfile", s.ctx, userID).Return(expectedProfile, nil)

	result, err := s.usecase.GetPublicProfile(s.ctx, userID)

	s.NoError(err)
	s.Equal(expectedProfile, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestFindMentors_Success() {
	topics := []string{"Technology Skills", "Career Guidance"}
	limit := 10
	offset := 0

	expectedProfiles := []userpkg.PublicProfile{
		{
			ID:                    primitive.NewObjectID(),
			DisplayName:           "TechMentor",
			IsMentor:              true,
			MentorshipTopics:      []string{"Technology Skills"},
			AvailableForMentoring: true,
		},
		{
			ID:                    primitive.NewObjectID(),
			DisplayName:           "CareerMentor",
			IsMentor:              true,
			MentorshipTopics:      []string{"Career Guidance"},
			AvailableForMentoring: true,
		},
	}

	s.mockUserRepo.On("FindMentors", s.ctx, topics, limit, offset).Return(expectedProfiles, nil)

	result, err := s.usecase.FindMentors(s.ctx, topics, limit, offset)

	s.NoError(err)
	s.Equal(expectedProfiles, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestFindMentors_DefaultLimit() {
	topics := []string{"Technology Skills"}
	limit := 0 // Invalid limit
	offset := 0

	expectedProfiles := []userpkg.PublicProfile{}

	// Should use default limit of 20
	s.mockUserRepo.On("FindMentors", s.ctx, topics, 20, offset).Return(expectedProfiles, nil)

	result, err := s.usecase.FindMentors(s.ctx, topics, limit, offset)

	s.NoError(err)
	s.Equal(expectedProfiles, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestFindMentees_Success() {
	topics := []string{"Study Techniques"}
	limit := 5
	offset := 0

	expectedProfiles := []userpkg.PublicProfile{
		{
			ID:               primitive.NewObjectID(),
			DisplayName:      "StudySeeker",
			IsMentee:         true,
			MentorshipTopics: []string{"Study Techniques"},
		},
	}

	s.mockUserRepo.On("FindMentees", s.ctx, topics, limit, offset).Return(expectedProfiles, nil)

	result, err := s.usecase.FindMentees(s.ctx, topics, limit, offset)

	s.NoError(err)
	s.Equal(expectedProfiles, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestSearchUsersByTopic_Success() {
	topic := "Technology Skills"
	isMentor := true
	limit := 10
	offset := 0

	expectedProfiles := []userpkg.PublicProfile{
		{
			ID:                    primitive.NewObjectID(),
			DisplayName:           "TechExpert",
			IsMentor:              true,
			MentorshipTopics:      []string{"Technology Skills"},
			AvailableForMentoring: true,
		},
	}

	s.mockUserRepo.On("SearchUsersByTopic", s.ctx, topic, isMentor, limit, offset).Return(expectedProfiles, nil)

	result, err := s.usecase.SearchUsersByTopic(s.ctx, topic, isMentor, limit, offset)

	s.NoError(err)
	s.Equal(expectedProfiles, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestSearchUsersByTopic_EmptyTopic() {
	topic := ""
	isMentor := true
	limit := 10
	offset := 0

	result, err := s.usecase.SearchUsersByTopic(s.ctx, topic, isMentor, limit, offset)

	s.Error(err)
	s.Equal("topic cannot be empty", err.Error())
	s.Nil(result)
}

func (s *userShareSpaceUsecaseTestSuite) TestGenerateDisplayName_Success() {
	baseName := "TechMentor"
	expectedName := "TechMentor"

	s.mockUserRepo.On("ExistsByDisplayName", s.ctx, expectedName).Return(false, nil)

	result, err := s.usecase.GenerateDisplayName(s.ctx, baseName)

	s.NoError(err)
	s.Equal(expectedName, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestGenerateDisplayName_WithNumber() {
	baseName := "TechMentor"
	expectedName := "TechMentor1"

	// First call returns true (exists), second call returns false (doesn't exist)
	s.mockUserRepo.On("ExistsByDisplayName", s.ctx, "TechMentor").Return(true, nil)
	s.mockUserRepo.On("ExistsByDisplayName", s.ctx, expectedName).Return(false, nil)

	result, err := s.usecase.GenerateDisplayName(s.ctx, baseName)

	s.NoError(err)
	s.Equal(expectedName, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestGenerateDisplayName_EmptyBaseName() {
	baseName := ""
	expectedName := "User"

	s.mockUserRepo.On("ExistsByDisplayName", s.ctx, expectedName).Return(false, nil)

	result, err := s.usecase.GenerateDisplayName(s.ctx, baseName)

	s.NoError(err)
	s.Equal(expectedName, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestGenerateDisplayName_CleanSpecialCharacters() {
	baseName := "Tech@Mentor#123!"
	expectedName := "TechMentor123"

	s.mockUserRepo.On("ExistsByDisplayName", s.ctx, expectedName).Return(false, nil)

	result, err := s.usecase.GenerateDisplayName(s.ctx, baseName)

	s.NoError(err)
	s.Equal(expectedName, result)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *userShareSpaceUsecaseTestSuite) TestGetAvailableMentorshipTopics() {
	topics := s.usecase.GetAvailableMentorshipTopics()

	s.NotEmpty(topics)
	s.Contains(topics, "Academic Support")
	s.Contains(topics, "Career Guidance")
	s.Contains(topics, "Technology Skills")
	s.Contains(topics, "Mental Health & Wellness")
}

func (s *userShareSpaceUsecaseTestSuite) TearDownTest() {
	s.mockUserRepo.AssertExpectations(s.T())
}
