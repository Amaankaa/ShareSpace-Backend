package repositories_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userShareSpaceTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	collection *mongo.Collection
	repo       *repositories.UserRepository
}

func TestUserShareSpaceTestSuite(t *testing.T) {
	suite.Run(t, new(userShareSpaceTestSuite))
}

func (s *userShareSpaceTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_sharespace_db")
	s.collection = s.db.Collection("users")
	s.repo = repositories.NewUserRepository(s.collection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)
}

func (s *userShareSpaceTestSuite) TearDownSuite() {
	// Create a new context for cleanup since the original might be cancelled
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	err := s.client.Database("test_sharespace_db").Drop(cleanupCtx)
	s.Require().NoError(err)

	_ = s.client.Disconnect(cleanupCtx)
	s.cancel()
}

func (s *userShareSpaceTestSuite) SetupTest() {
	// Clean the collection before each test
	_, err := s.collection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

// Helper function to create a test user with ShareSpace fields
func (s *userShareSpaceTestSuite) createTestUser(displayName string, isMentor, isMentee bool, topics []string) userpkg.User {
	user := userpkg.User{
		ID:         primitive.NewObjectID(),
		Username:   "testuser_" + displayName,
		Fullname:   "Test User " + displayName,
		Email:      displayName + "@test.com",
		Password:   "hashedpassword",
		Role:       "user",
		IsVerified: true,
		Bio:        "Test bio",
		UpdatedAt:  time.Now(),

		// ShareSpace fields
		DisplayName:           displayName,
		IsAnonymous:           true,
		IsMentor:              isMentor,
		IsMentee:              isMentee,
		MentorshipTopics:      topics,
		MentorshipBio:         "I can help with " + displayName + " topics",
		AvailableForMentoring: isMentor,
		PrivacySettings: userpkg.PrivacySettings{
			ShowRealName:       false,
			ShowProfilePicture: false,
			ShowContactInfo:    false,
		},
	}

	_, err := s.collection.InsertOne(s.ctx, user)
	s.Require().NoError(err)

	return user
}

func (s *userShareSpaceTestSuite) TestExistsByDisplayName() {
	// Create a user with a display name
	s.createTestUser("TechMentor123", true, false, []string{"Technology Skills"})

	// Test existing display name
	exists, err := s.repo.ExistsByDisplayName(s.ctx, "TechMentor123")
	s.Require().NoError(err)
	s.True(exists)

	// Test non-existing display name
	exists, err = s.repo.ExistsByDisplayName(s.ctx, "NonExistentName")
	s.Require().NoError(err)
	s.False(exists)
}

func (s *userShareSpaceTestSuite) TestGetPublicProfile() {
	// Create a test user
	user := s.createTestUser("StudyBuddy", true, true, []string{"Study Techniques", "Time Management"})

	// Get public profile
	profile, err := s.repo.GetPublicProfile(s.ctx, user.ID.Hex())
	s.Require().NoError(err)

	// Verify public profile fields
	s.Equal(user.ID, profile.ID)
	s.Equal("StudyBuddy", profile.DisplayName)
	s.Equal("Test bio", profile.Bio)
	s.True(profile.IsMentor)
	s.True(profile.IsMentee)
	s.Equal([]string{"Study Techniques", "Time Management"}, profile.MentorshipTopics)
	s.Equal("I can help with StudyBuddy topics", profile.MentorshipBio)
	s.True(profile.AvailableForMentoring)

	// Verify privacy - real name should not be shown
	s.Empty(profile.Fullname)
	s.Empty(profile.ContactInfo.Phone)
}

func (s *userShareSpaceTestSuite) TestFindMentors() {
	// Create test mentors
	s.createTestUser("TechMentor", true, false, []string{"Technology Skills", "Career Guidance"})
	s.createTestUser("StudyMentor", true, false, []string{"Study Techniques", "Time Management"})
	s.createTestUser("CareerMentor", true, false, []string{"Career Guidance", "Interview Preparation"})

	// Create a mentee (should not appear in mentor search)
	s.createTestUser("StudentSeeker", false, true, []string{"Technology Skills"})

	// Search for mentors by topic
	mentors, err := s.repo.FindMentors(s.ctx, []string{"Technology Skills"}, 10, 0)
	s.Require().NoError(err)
	s.Len(mentors, 1)
	s.Equal("TechMentor", mentors[0].DisplayName)

	// Search for mentors by multiple topics
	mentors, err = s.repo.FindMentors(s.ctx, []string{"Career Guidance"}, 10, 0)
	s.Require().NoError(err)
	s.Len(mentors, 2) // TechMentor and CareerMentor

	// Test pagination
	mentors, err = s.repo.FindMentors(s.ctx, []string{"Career Guidance"}, 1, 0)
	s.Require().NoError(err)
	s.Len(mentors, 1)

	mentors, err = s.repo.FindMentors(s.ctx, []string{"Career Guidance"}, 1, 1)
	s.Require().NoError(err)
	s.Len(mentors, 1)
}

func (s *userShareSpaceTestSuite) TestFindMentees() {
	// Create test mentees
	s.createTestUser("TechStudent", false, true, []string{"Technology Skills", "Career Guidance"})
	s.createTestUser("StudyStudent", false, true, []string{"Study Techniques", "Time Management"})

	// Create a mentor (should not appear in mentee search)
	s.createTestUser("ExpertMentor", true, false, []string{"Technology Skills"})

	// Search for mentees by topic
	mentees, err := s.repo.FindMentees(s.ctx, []string{"Technology Skills"}, 10, 0)
	s.Require().NoError(err)
	s.Len(mentees, 1)
	s.Equal("TechStudent", mentees[0].DisplayName)

	// Search for mentees by multiple topics
	mentees, err = s.repo.FindMentees(s.ctx, []string{"Study Techniques", "Time Management"}, 10, 0)
	s.Require().NoError(err)
	s.Len(mentees, 1)
	s.Equal("StudyStudent", mentees[0].DisplayName)
}

func (s *userShareSpaceTestSuite) TestSearchUsersByTopic() {
	// Create test users
	s.createTestUser("TechMentor", true, false, []string{"Technology Skills"})
	s.createTestUser("TechStudent", false, true, []string{"Technology Skills"})
	s.createTestUser("StudyMentor", true, false, []string{"Study Techniques"})

	// Search for mentors in Technology Skills
	users, err := s.repo.SearchUsersByTopic(s.ctx, "Technology Skills", true, 10, 0)
	s.Require().NoError(err)
	s.Len(users, 1)
	s.Equal("TechMentor", users[0].DisplayName)

	// Search for mentees in Technology Skills
	users, err = s.repo.SearchUsersByTopic(s.ctx, "Technology Skills", false, 10, 0)
	s.Require().NoError(err)
	s.Len(users, 1)
	s.Equal("TechStudent", users[0].DisplayName)

	// Search for non-existent topic
	users, err = s.repo.SearchUsersByTopic(s.ctx, "Non-existent Topic", true, 10, 0)
	s.Require().NoError(err)
	s.Len(users, 0)
}

func (s *userShareSpaceTestSuite) TestPrivacySettings() {
	// Create user with privacy settings that allow showing real name
	user := s.createTestUser("OpenUser", true, false, []string{"Career Guidance"})

	// Update privacy settings to show real name
	_, err := s.collection.UpdateOne(
		s.ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"privacySettings.showRealName": true}},
	)
	s.Require().NoError(err)

	// Get public profile - should now show real name
	profile, err := s.repo.GetPublicProfile(s.ctx, user.ID.Hex())
	s.Require().NoError(err)
	s.Equal("Test User OpenUser", profile.Fullname)
}
