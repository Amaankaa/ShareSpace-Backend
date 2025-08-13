package repositories_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	collection *mongo.Collection
	repo       *repositories.UserRepository
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

func (s *userRepositoryTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_blog_db")
	s.collection = s.db.Collection(testUserCollection)
	s.repo = repositories.NewUserRepository(s.collection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *userRepositoryTestSuite) TearDownSuite() {
	s.collection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *userRepositoryTestSuite) SetupTest() {
	_, err := s.collection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

func (s *userRepositoryTestSuite) TestCreateUser() {
	user := userpkg.User{
		Username: "testuser",
		Password: "hashed-pass",
		Email:    "test@example.com",
		Fullname: "Test User",
	}

	created, err := s.repo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	s.NotEmpty(created.ID)
	s.Equal("testuser", created.Username)
	s.Equal("test@example.com", created.Email)
	s.Equal("", created.Password) // password scrubbed
	s.False(created.IsVerified)
}

func (s *userRepositoryTestSuite) TestExistsByUsername() {
	_, err := s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "checkuser",
		Password: "secret",
		Email:    "check@example.com",
		Fullname: "Checker",
	})
	s.Require().NoError(err)

	found, err := s.repo.ExistsByUsername(s.ctx, "checkuser")
	s.Require().NoError(err)
	s.True(found)

	notFound, err := s.repo.ExistsByUsername(s.ctx, "ghost")
	s.Require().NoError(err)
	s.False(notFound)
}

func (s *userRepositoryTestSuite) TestExistsByEmail() {
	_, err := s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "emailer",
		Password: "secret",
		Email:    "mail@example.com",
		Fullname: "Mailer",
	})
	s.Require().NoError(err)

	found, err := s.repo.ExistsByEmail(s.ctx, "mail@example.com")
	s.Require().NoError(err)
	s.True(found)

	notFound, err := s.repo.ExistsByEmail(s.ctx, "notfound@example.com")
	s.Require().NoError(err)
	s.False(notFound)
}

func (s *userRepositoryTestSuite) TestCountUsers() {
	initial, err := s.repo.CountUsers(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(0), initial)

	_, err = s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "counter",
		Password: "pw",
		Email:    "count@example.com",
		Fullname: "Counter",
	})
	s.Require().NoError(err)

	count, err := s.repo.CountUsers(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}

// TestUpdateUserRoleByID_Success updates the role to admin
func (s *userRepositoryTestSuite) TestUpdateUserRoleByID_Success() {
	// Arrange: create user with initial role
	user := userpkg.User{
		Username: "roleuser",
		Password: "pass",
		Email:    "role@example.com",
		Fullname: "Role User",
		Role:     "user",
	}
	created, err := s.repo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	id := created.ID.Hex()

	// Act: promote to admin
	err = s.repo.UpdateUserRoleByID(s.ctx, id, "admin")
	// Assert
	s.NoError(err)
	updated, err := s.repo.FindByID(s.ctx, id)
	s.Require().NoError(err)
	s.Equal("admin", updated.Role)
}

// TestUpdateUserRoleByID_NotFound returns error when user does not exist
func (s *userRepositoryTestSuite) TestUpdateUserRoleByID_NotFound() {
	fakeID := primitive.NewObjectID().Hex()
	err := s.repo.UpdateUserRoleByID(s.ctx, fakeID, "admin")
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

// TestUpdateIsVerifiedByEmail_Success sets isVerified to true
func (s *userRepositoryTestSuite) TestUpdateIsVerifiedByEmail_Success() {
	// Arrange: create unverified user
	usr, err := s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "verifyuser",
		Password: "pass",
		Email:    "verify@example.com",
		Fullname: "Verify User",
	})
	s.Require().NoError(err)
	error := s.repo.UpdateIsVerifiedByEmail(s.ctx, "verify@example.com", true)
	s.NoError(error)

	// Assert in DB
	updated, err := s.repo.FindByID(s.ctx, usr.ID.Hex())
	s.Require().NoError(err)
	s.True(updated.IsVerified)
}

// TestUpdateIsVerifiedByEmail_NotFound returns error for missing user
func (s *userRepositoryTestSuite) TestUpdateIsVerifiedByEmail_NotFound() {
	err := s.repo.UpdateIsVerifiedByEmail(s.ctx, "notfound@example.com", true)
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

func (s *userRepositoryTestSuite) TestUpdateProfile_Success() {
	// Arrange: create a user
	user := userpkg.User{
		Username:       "profileuser",
		Password:       "pass",
		Email:          "profile@example.com",
		Fullname:       "Profile User",
		Bio:            "Old bio",
		ProfilePicture: "oldpic.jpg",
		ContactInfo:    userpkg.ContactInfo{Phone: "123", Website: "oldsite.com"},
	}
	created, err := s.repo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	id := created.ID.Hex()

	// Act: update profile fields
	updates := userpkg.UpdateProfileRequest{
		Fullname:       "New Name",
		Bio:            "New bio",
		ProfilePicture: "newpic.jpg",
		ContactInfo:    userpkg.ContactInfo{Phone: "456", Website: "newsite.com", Twitter: "@new"},
	}
	updated, err := s.repo.UpdateProfile(s.ctx, id, updates)
	s.Require().NoError(err)

	// Assert: fields are updated
	s.Equal("New Name", updated.Fullname)
	s.Equal("New bio", updated.Bio)
	s.Equal("newpic.jpg", updated.ProfilePicture)
	s.Equal(userpkg.ContactInfo{Phone: "456", Website: "newsite.com", Twitter: "@new"}, updated.ContactInfo)
	s.WithinDuration(time.Now(), updated.UpdatedAt, time.Second*2)
}

// TestUpdateProfile_NotFound returns error when user does not exist
func (s *userRepositoryTestSuite) TestUpdateProfile_NotFound() {
	fakeID := primitive.NewObjectID().Hex()
	updates := userpkg.UpdateProfileRequest{
		Fullname: "Ghost",
	}
	_, err := s.repo.UpdateProfile(s.ctx, fakeID, updates)
	s.Error(err)
}

// TestUpdateProfile_InvalidID returns error for invalid object ID
func (s *userRepositoryTestSuite) TestUpdateProfile_InvalidID() {
	invalidID := "not-a-valid-hex"
	updates := userpkg.UpdateProfileRequest{
		Fullname: "Invalid",
	}
	_, err := s.repo.UpdateProfile(s.ctx, invalidID, updates)
	s.Error(err)
}

func (s *userRepositoryTestSuite) TestGetUserProfile_Success() {
	// Arrange: create a user
	user := userpkg.User{
		Username:       "profileuser",
		Password:       "pass",
		Email:          "profile@example.com",
		Fullname:       "Profile User",
		Bio:            "User bio",
		ProfilePicture: "profile.jpg",
		ContactInfo: userpkg.ContactInfo{
			Phone:   "1234567890",
			Website: "example.com",
		},
	}
	created, err := s.repo.CreateUser(s.ctx, user)
	s.Require().NoError(err, "Failed to create user")
	id := created.ID.Hex()

	// Act
	got, err := s.repo.GetUserProfile(s.ctx, id)

	// Assert
	s.NoError(err, "Error occurred while fetching user profile")
	s.Equal(created.ID, got.ID, "User ID mismatch")
	s.Equal(created.Username, got.Username, "Username mismatch")
	s.Equal(created.Email, got.Email, "Email mismatch")
	s.Equal(created.Fullname, got.Fullname, "Fullname mismatch")
	s.Equal(created.Bio, got.Bio, "Bio mismatch")
	s.Equal(created.ProfilePicture, got.ProfilePicture, "Profile picture mismatch")
	s.Equal(created.ContactInfo, got.ContactInfo, "Contact info mismatch")
}

func (s *userRepositoryTestSuite) TestGetUserProfile_NotFound() {
	// Arrange: generate a non-existent user ID
	nonExistentID := primitive.NewObjectID().Hex()

	// Act
	got, err := s.repo.GetUserProfile(s.ctx, nonExistentID)

	// Assert
	s.Error(err, "Expected error for non-existent user ID")
	s.Contains(err.Error(), "no documents in result", "Error message mismatch")
	s.Equal(userpkg.User{}, got, "Expected empty user object")
}
