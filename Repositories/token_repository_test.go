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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const testTokenCollection = "test_tokens"

type tokenRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	collection *mongo.Collection
	repo       *repositories.TokenRepository
}

func TestTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(tokenRepositoryTestSuite))
}

func (s *tokenRepositoryTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_blog_db")
	s.collection = s.db.Collection(testTokenCollection)
	s.repo = repositories.NewTokenRepository(s.collection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *tokenRepositoryTestSuite) TearDownSuite() {
	s.collection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *tokenRepositoryTestSuite) SetupTest() {
	_, err := s.collection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

func (s *tokenRepositoryTestSuite) TestStoreToken() {
	assert := assert.New(s.T())

	userID := primitive.NewObjectID()
	token := userpkg.Token{
		UserID:       userID,
		AccessToken:  "access-123",
		RefreshToken: "refresh-123",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err := s.repo.StoreToken(s.ctx, token)
	assert.NoError(err)

	var found userpkg.Token
	err = s.collection.FindOne(s.ctx, bson.M{"refresh_token": token.RefreshToken}).Decode(&found)
	assert.NoError(err)
	assert.Equal(token.UserID, found.UserID)
	assert.Equal(token.AccessToken, found.AccessToken)
	assert.Equal(token.RefreshToken, found.RefreshToken)
	assert.WithinDuration(token.CreatedAt, found.CreatedAt, time.Second)
	assert.WithinDuration(token.ExpiresAt, found.ExpiresAt, time.Second)
}

func (s *tokenRepositoryTestSuite) TestFindByRefreshToken() {
	assert := assert.New(s.T())

	token := userpkg.Token{
		UserID:       primitive.NewObjectID(),
		AccessToken:  "acc-token",
		RefreshToken: "find-this-token",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(12 * time.Hour),
	}
	s.Require().NoError(s.repo.StoreToken(s.ctx, token))

	found, err := s.repo.FindByRefreshToken(s.ctx, "find-this-token")
	assert.NoError(err)
	assert.Equal(token.RefreshToken, found.RefreshToken)
	assert.Equal(token.AccessToken, found.AccessToken)
	assert.Equal(token.UserID, found.UserID)

	// Non-existent token
	_, err = s.repo.FindByRefreshToken(s.ctx, "ghost-token")
	assert.Error(err)
	assert.Equal(mongo.ErrNoDocuments, err)
}

func (s *tokenRepositoryTestSuite) TestDeleteByRefreshToken() {
	assert := assert.New(s.T())

	token := userpkg.Token{
		UserID:       primitive.NewObjectID(),
		AccessToken:  "to-delete-access",
		RefreshToken: "to-delete-refresh",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	s.Require().NoError(s.repo.StoreToken(s.ctx, token))

	err := s.repo.DeleteByRefreshToken(s.ctx, "to-delete-refresh")
	assert.NoError(err)

	_, err = s.repo.FindByRefreshToken(s.ctx, "to-delete-refresh")
	assert.Error(err)
	assert.Equal(mongo.ErrNoDocuments, err)
}

func (s *tokenRepositoryTestSuite) TestDeleteTokensByUserID() {
	assert := assert.New(s.T())

	// Arrange
	userID := primitive.NewObjectID()

	tokens := []interface{}{
		userpkg.Token{
			UserID:       userID,
			AccessToken:  "access-token-1",
			RefreshToken: "refresh-token-1",
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		},
		userpkg.Token{
			UserID:       userID,
			AccessToken:  "access-token-2",
			RefreshToken: "refresh-token-2",
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		},
	}

	_, err := s.collection.InsertMany(s.ctx, tokens)
	assert.NoError(err)

	// Make sure tokens exist before delete
	countBefore, err := s.collection.CountDocuments(s.ctx, bson.M{"user_id": userID})
	assert.NoError(err)
	assert.Equal(int64(2), countBefore)

	// Act
	err = s.repo.DeleteTokensByUserID(s.ctx, userID.Hex())
	assert.NoError(err)

	// Assert
	countAfter, err := s.collection.CountDocuments(s.ctx, bson.M{"user_id": userID})
	assert.NoError(err)
	assert.Equal(int64(0), countAfter)
}