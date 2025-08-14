package repositories_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mentorshipRepositoryTestSuite struct {
	suite.Suite
	db                    *mongo.Database
	client                *mongo.Client
	ctx                   context.Context
	cancel                context.CancelFunc
	requestsCollection    *mongo.Collection
	connectionsCollection *mongo.Collection
	repo                  *repositories.MentorshipRepository
}

func TestMentorshipRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(mentorshipRepositoryTestSuite))
}

func (s *mentorshipRepositoryTestSuite) SetupSuite() {
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
	s.db = client.Database("test_mentorship_db")
	s.requestsCollection = s.db.Collection("mentorship_requests")
	s.connectionsCollection = s.db.Collection("mentorship_connections")
	s.repo = repositories.NewMentorshipRepository(s.requestsCollection, s.connectionsCollection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 60*time.Second)
}

func (s *mentorshipRepositoryTestSuite) TearDownSuite() {
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	err := s.client.Database("test_mentorship_db").Drop(cleanupCtx)
	s.Require().NoError(err)

	_ = s.client.Disconnect(cleanupCtx)
	s.cancel()
}

func (s *mentorshipRepositoryTestSuite) SetupTest() {
	// Clean collections before each test
	_, err := s.requestsCollection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
	_, err = s.connectionsCollection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

// Helper function to create test mentorship request
func (s *mentorshipRepositoryTestSuite) createTestRequest(menteeID, mentorID primitive.ObjectID, topics []string) mentorshippkg.MentorshipRequest {
	request := mentorshippkg.MentorshipRequest{
		MenteeID:  menteeID,
		MentorID:  mentorID,
		Status:    mentorshippkg.StatusPending,
		Message:   "I would like to learn from you",
		Topics:    topics,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdRequest, err := s.repo.CreateRequest(s.ctx, request)
	s.Require().NoError(err)
	return createdRequest
}

// Helper function to create test mentorship connection
func (s *mentorshipRepositoryTestSuite) createTestConnection(menteeID, mentorID, requestID primitive.ObjectID, topics []string) mentorshippkg.MentorshipConnection {
	connection := mentorshippkg.MentorshipConnection{
		MenteeID:  menteeID,
		MentorID:  mentorID,
		RequestID: requestID,
		Status:    mentorshippkg.ConnectionActive,
		Topics:    topics,
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdConnection, err := s.repo.CreateConnection(s.ctx, connection)
	s.Require().NoError(err)
	return createdConnection
}

func (s *mentorshipRepositoryTestSuite) TestCreateRequest() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	topics := []string{"Technology Skills", "Career Guidance"}

	request := mentorshippkg.MentorshipRequest{
		MenteeID: menteeID,
		MentorID: mentorID,
		Message:  "I would like to learn from you",
		Topics:   topics,
	}

	createdRequest, err := s.repo.CreateRequest(s.ctx, request)

	s.NoError(err)
	s.NotEmpty(createdRequest.ID)
	s.Equal(menteeID, createdRequest.MenteeID)
	s.Equal(mentorID, createdRequest.MentorID)
	s.Equal(mentorshippkg.StatusPending, createdRequest.Status)
	s.Equal("I would like to learn from you", createdRequest.Message)
	s.Equal(topics, createdRequest.Topics)
	s.NotZero(createdRequest.CreatedAt)
	s.NotZero(createdRequest.UpdatedAt)
}

func (s *mentorshipRepositoryTestSuite) TestGetRequestByID() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	topics := []string{"Study Techniques"}

	createdRequest := s.createTestRequest(menteeID, mentorID, topics)

	retrievedRequest, err := s.repo.GetRequestByID(s.ctx, createdRequest.ID.Hex())

	s.NoError(err)
	s.Equal(createdRequest.ID, retrievedRequest.ID)
	s.Equal(createdRequest.MenteeID, retrievedRequest.MenteeID)
	s.Equal(createdRequest.MentorID, retrievedRequest.MentorID)
	s.Equal(createdRequest.Status, retrievedRequest.Status)
}

func (s *mentorshipRepositoryTestSuite) TestGetRequestByID_NotFound() {
	nonExistentID := primitive.NewObjectID()

	_, err := s.repo.GetRequestByID(s.ctx, nonExistentID.Hex())

	s.Error(err)
	s.Equal(mongo.ErrNoDocuments, err)
}

func (s *mentorshipRepositoryTestSuite) TestExistsPendingRequest() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	topics := []string{"Technology Skills"}

	// Test when no request exists
	exists, err := s.repo.ExistsPendingRequest(s.ctx, menteeID.Hex(), mentorID.Hex())
	s.NoError(err)
	s.False(exists)

	// Create a pending request
	s.createTestRequest(menteeID, mentorID, topics)

	// Test when pending request exists
	exists, err = s.repo.ExistsPendingRequest(s.ctx, menteeID.Hex(), mentorID.Hex())
	s.NoError(err)
	s.True(exists)
}

func (s *mentorshipRepositoryTestSuite) TestUpdateRequestStatus() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	topics := []string{"Career Guidance"}

	createdRequest := s.createTestRequest(menteeID, mentorID, topics)

	err := s.repo.UpdateRequestStatus(s.ctx, createdRequest.ID.Hex(), mentorshippkg.StatusAccepted)
	s.NoError(err)

	// Verify the status was updated
	updatedRequest, err := s.repo.GetRequestByID(s.ctx, createdRequest.ID.Hex())
	s.NoError(err)
	s.Equal(mentorshippkg.StatusAccepted, updatedRequest.Status)
	s.NotNil(updatedRequest.ResponsedAt)
}

func (s *mentorshipRepositoryTestSuite) TestGetRequestsByMentee() {
	menteeID := primitive.NewObjectID()
	mentor1ID := primitive.NewObjectID()
	mentor2ID := primitive.NewObjectID()
	topics := []string{"Technology Skills"}

	// Create multiple requests for the same mentee
	s.createTestRequest(menteeID, mentor1ID, topics)
	s.createTestRequest(menteeID, mentor2ID, topics)

	// Create a request for a different mentee (should not be returned)
	otherMenteeID := primitive.NewObjectID()
	s.createTestRequest(otherMenteeID, mentor1ID, topics)

	requests, err := s.repo.GetRequestsByMentee(s.ctx, menteeID.Hex(), 10, 0)

	s.NoError(err)
	s.Len(requests, 2)
	for _, request := range requests {
		s.Equal(menteeID, request.MenteeID)
	}
}

func (s *mentorshipRepositoryTestSuite) TestGetPendingRequestsByMentor() {
	mentorID := primitive.NewObjectID()
	mentee1ID := primitive.NewObjectID()
	mentee2ID := primitive.NewObjectID()
	topics := []string{"Study Techniques"}

	// Create pending requests
	request1 := s.createTestRequest(mentee1ID, mentorID, topics)
	s.createTestRequest(mentee2ID, mentorID, topics)

	// Accept one request (should not appear in pending)
	err := s.repo.UpdateRequestStatus(s.ctx, request1.ID.Hex(), mentorshippkg.StatusAccepted)
	s.Require().NoError(err)

	pendingRequests, err := s.repo.GetPendingRequestsByMentor(s.ctx, mentorID.Hex(), 10, 0)

	s.NoError(err)
	s.Len(pendingRequests, 1)
	s.Equal(mentee2ID, pendingRequests[0].MenteeID)
	s.Equal(mentorshippkg.StatusPending, pendingRequests[0].Status)
}

func (s *mentorshipRepositoryTestSuite) TestCreateConnection() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()
	topics := []string{"Technology Skills", "Career Guidance"}

	connection := mentorshippkg.MentorshipConnection{
		MenteeID:  menteeID,
		MentorID:  mentorID,
		RequestID: requestID,
		Topics:    topics,
	}

	createdConnection, err := s.repo.CreateConnection(s.ctx, connection)

	s.NoError(err)
	s.NotEmpty(createdConnection.ID)
	s.Equal(menteeID, createdConnection.MenteeID)
	s.Equal(mentorID, createdConnection.MentorID)
	s.Equal(requestID, createdConnection.RequestID)
	s.Equal(mentorshippkg.ConnectionActive, createdConnection.Status)
	s.Equal(topics, createdConnection.Topics)
	s.NotZero(createdConnection.StartedAt)
	s.NotZero(createdConnection.CreatedAt)
	s.NotZero(createdConnection.UpdatedAt)
}

func (s *mentorshipRepositoryTestSuite) TestGetActiveConnectionsByUser() {
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()
	topics := []string{"Technology Skills"}

	// Create active connection where user is mentee
	s.createTestConnection(userID, otherUserID, requestID, topics)

	// Create active connection where user is mentor
	s.createTestConnection(otherUserID, userID, requestID, topics)

	// Create connection for different user (should not be returned)
	differentUserID := primitive.NewObjectID()
	s.createTestConnection(differentUserID, otherUserID, requestID, topics)

	connections, err := s.repo.GetActiveConnectionsByUser(s.ctx, userID.Hex())

	s.NoError(err)
	s.Len(connections, 2)

	// Verify user is either mentee or mentor in all returned connections
	for _, conn := range connections {
		s.True(conn.MenteeID == userID || conn.MentorID == userID)
		s.Equal(mentorshippkg.ConnectionActive, conn.Status)
	}
}

func (s *mentorshipRepositoryTestSuite) TestUpdateLastInteraction() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()
	topics := []string{"Study Techniques"}

	connection := s.createTestConnection(menteeID, mentorID, requestID, topics)

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	err := s.repo.UpdateLastInteraction(s.ctx, connection.ID.Hex())
	s.NoError(err)

	// Verify last interaction was updated
	updatedConnection, err := s.repo.GetConnectionByID(s.ctx, connection.ID.Hex())
	s.NoError(err)
	s.NotNil(updatedConnection.LastInteraction)
	s.True(updatedConnection.LastInteraction.After(connection.StartedAt))
}

func (s *mentorshipRepositoryTestSuite) TestEndConnection() {
	menteeID := primitive.NewObjectID()
	mentorID := primitive.NewObjectID()
	requestID := primitive.NewObjectID()
	topics := []string{"Career Guidance"}

	connection := s.createTestConnection(menteeID, mentorID, requestID, topics)

	rating := 5
	feedback := "Great mentorship experience!"
	endReason := "Goals achieved"

	err := s.repo.EndConnection(s.ctx, connection.ID.Hex(), endReason, &rating, feedback, false)
	s.NoError(err)

	// Verify connection was ended
	endedConnection, err := s.repo.GetConnectionByID(s.ctx, connection.ID.Hex())
	s.NoError(err)
	s.Equal(mentorshippkg.ConnectionEnded, endedConnection.Status)
	s.Equal(endReason, endedConnection.EndReason)
	s.NotNil(endedConnection.EndedAt)
	s.Equal(&rating, endedConnection.MenteeRating) // Ended by mentee
	s.Equal(feedback, endedConnection.MenteeFeedback)
}

func (s *mentorshipRepositoryTestSuite) TestGetMentorshipStats() {
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	topics := []string{"Technology Skills"}

	// Create some test data
	// As mentor: 2 requests, 1 accepted, 1 active connection
	request1 := s.createTestRequest(otherUserID, userID, topics)
	s.createTestRequest(otherUserID, userID, topics) // Second request stays pending

	err := s.repo.UpdateRequestStatus(s.ctx, request1.ID.Hex(), mentorshippkg.StatusAccepted)
	s.Require().NoError(err)

	s.createTestConnection(otherUserID, userID, request1.ID, topics)

	// As mentee: 1 request, 1 accepted
	request3 := s.createTestRequest(userID, otherUserID, topics)
	err = s.repo.UpdateRequestStatus(s.ctx, request3.ID.Hex(), mentorshippkg.StatusAccepted)
	s.Require().NoError(err)

	stats, err := s.repo.GetMentorshipStats(s.ctx, userID.Hex())

	s.NoError(err)
	s.Equal(userID, stats.UserID)
	s.Equal(2, stats.TotalMentorRequests)
	s.Equal(1, stats.AcceptedMentorRequests)
	s.Equal(1, stats.ActiveMentorships)
	s.Equal(1, stats.TotalMenteeRequests)
	s.Equal(1, stats.AcceptedMenteeRequests)
}
