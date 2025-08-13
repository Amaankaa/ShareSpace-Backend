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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testResetCollection = "test_password_resets"
const testUserCollection = "test_users"

type passwordResetRepoTestSuite struct {
	suite.Suite
	db              *mongo.Database
	resetRepo       *repositories.PasswordResetRepo
	resetCollection *mongo.Collection
	userCollection  *mongo.Collection
	ctx             context.Context
	cancel          context.CancelFunc
	client          *mongo.Client
}

func TestPasswordResetRepoTestSuite(t *testing.T) {
	suite.Run(t, new(passwordResetRepoTestSuite))
}

func (s *passwordResetRepoTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_blog_db")
	s.resetCollection = s.db.Collection(testResetCollection)
	s.userCollection = s.db.Collection(testUserCollection)
	s.resetRepo = repositories.NewPasswordResetRepo(s.resetCollection, s.userCollection)
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *passwordResetRepoTestSuite) TearDownSuite() {
	s.resetCollection.Drop(s.ctx)
	s.userCollection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *passwordResetRepoTestSuite) SetupTest() {
	s.resetCollection.DeleteMany(s.ctx, bson.M{})
	s.userCollection.DeleteMany(s.ctx, bson.M{})
}

func (s *passwordResetRepoTestSuite) TestStoreAndGetResetRequest() {
	assert := assert.New(s.T())

	expiration := time.Now().Add(10 * time.Minute)

	reset := userpkg.PasswordReset{
		Email:        "test@example.com",
		OTP:          "123456",
		ExpiresAt:    expiration,
		AttemptCount: 0,
	}

	err := s.resetRepo.StoreResetRequest(s.ctx, reset)
	assert.NoError(err)

	found, err := s.resetRepo.GetResetRequest(s.ctx, reset.Email)
	assert.NoError(err)
	assert.Equal(reset.Email, found.Email)
	assert.Equal(reset.OTP, found.OTP)
	assert.Equal(reset.AttemptCount, found.AttemptCount)
	assert.WithinDuration(reset.ExpiresAt, found.ExpiresAt, time.Second)
}

func (s *passwordResetRepoTestSuite) TestIncrementAttemptCount() {
	assert := assert.New(s.T())

	reset := userpkg.PasswordReset{
		Email:        "test@example.com",
		OTP:          "123456",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 1,
	}
	s.resetRepo.StoreResetRequest(s.ctx, reset)

	err := s.resetRepo.IncrementAttemptCount(s.ctx, reset.Email)
	assert.NoError(err)

	updated, err := s.resetRepo.GetResetRequest(s.ctx, reset.Email)
	assert.NoError(err)
	assert.Equal(2, updated.AttemptCount)
}

func (s *passwordResetRepoTestSuite) TestDeleteResetRequest() {
	assert := assert.New(s.T())

	reset := userpkg.PasswordReset{
		Email:        "delete@example.com",
		OTP:          "000000",
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptCount: 0,
	}
	s.resetRepo.StoreResetRequest(s.ctx, reset)

	err := s.resetRepo.DeleteResetRequest(s.ctx, reset.Email)
	assert.NoError(err)

	_, err = s.resetRepo.GetResetRequest(s.ctx, reset.Email)
	assert.Error(err)
	assert.Equal(mongo.ErrNoDocuments, err)
}