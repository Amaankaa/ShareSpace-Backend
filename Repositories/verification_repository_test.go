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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testVerificationCollection = "test_verifications"

// verificationRepoTestSuite runs tests against a live MongoDB instance
// using a separate collection for verifications.
type verificationRepoTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	collection *mongo.Collection
	repo       *repositories.VerificationRepo
	ctx        context.Context
	cancel     context.CancelFunc
}

func TestVerificationRepoTestSuite(t *testing.T) {
	suite.Run(t, new(verificationRepoTestSuite))
}

func (s *verificationRepoTestSuite) SetupSuite() {
	// load env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	s.Require().NoError(err)
	s.client = client
	s.db = client.Database("test_blog_db")
	s.collection = s.db.Collection(testVerificationCollection)
	s.repo = repositories.NewVerificationRepo(s.collection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *verificationRepoTestSuite) TearDownSuite() {
	s.collection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *verificationRepoTestSuite) SetupTest() {
	// clear collection
	s.Require().NoError(s.collection.Drop(s.ctx))
}

func (s *verificationRepoTestSuite) TestStoreAndGetVerification() {
	v := userpkg.Verification{
		Email:        "test@example.com",
		OTP:          "otp123",
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptCount: 0,
	}
	// Store
	s.Require().NoError(s.repo.StoreVerification(s.ctx, v))

	// Get
	rcv, err := s.repo.GetVerification(s.ctx, v.Email)
	s.Require().NoError(err)
	s.Equal(v.Email, rcv.Email)
	s.Equal(v.OTP, rcv.OTP)
}

func (s *verificationRepoTestSuite) TestDeleteVerification() {
	email := "delete@example.com"
	v := userpkg.Verification{Email: email, OTP: "x", ExpiresAt: time.Now(), AttemptCount: 0}
	s.Require().NoError(s.repo.StoreVerification(s.ctx, v))

	s.Require().NoError(s.repo.DeleteVerification(s.ctx, email))

	_, err := s.repo.GetVerification(s.ctx, email)
	s.Error(err)
}

func (s *verificationRepoTestSuite) TestIncrementAttemptCount() {
	email := "inc@example.com"
	v := userpkg.Verification{Email: email, OTP: "x", ExpiresAt: time.Now(), AttemptCount: 0}
	s.Require().NoError(s.repo.StoreVerification(s.ctx, v))

	// increment
	s.Require().NoError(s.repo.IncrementAttemptCount(s.ctx, email))

	// check
	rcv, err := s.repo.GetVerification(s.ctx, email)
	s.Require().NoError(err)
	s.Equal(1, rcv.AttemptCount)
}

// TestIncrementAttemptCount_RecordNotFound ensures an error is returned when no verification exists
func (s *verificationRepoTestSuite) TestIncrementAttemptCount_RecordNotFound() {
	// attempt to increment on a non-existent email
	err := s.repo.IncrementAttemptCount(s.ctx, "missing@example.com")
	s.Error(err)
	s.Contains(err.Error(), "verification record not found")
}
