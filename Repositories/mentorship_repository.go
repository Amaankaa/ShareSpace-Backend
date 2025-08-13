package repositories

import (
	"context"
	"errors"
	"time"

	mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MentorshipRepository struct {
	requestsCollection    *mongo.Collection
	connectionsCollection *mongo.Collection
}

func NewMentorshipRepository(requestsCollection, connectionsCollection *mongo.Collection) *MentorshipRepository {
	return &MentorshipRepository{
		requestsCollection:    requestsCollection,
		connectionsCollection: connectionsCollection,
	}
}

// Mentorship Request operations

func (mr *MentorshipRepository) CreateRequest(ctx context.Context, request mentorshippkg.MentorshipRequest) (mentorshippkg.MentorshipRequest, error) {
	request.ID = primitive.NewObjectID()
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	request.Status = mentorshippkg.StatusPending

	_, err := mr.requestsCollection.InsertOne(ctx, request)
	if err != nil {
		return mentorshippkg.MentorshipRequest{}, err
	}

	return request, nil
}

func (mr *MentorshipRepository) GetRequestByID(ctx context.Context, requestID string) (mentorshippkg.MentorshipRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return mentorshippkg.MentorshipRequest{}, errors.New("invalid request ID")
	}

	var request mentorshippkg.MentorshipRequest
	err = mr.requestsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&request)
	if err != nil {
		return mentorshippkg.MentorshipRequest{}, err
	}

	return request, nil
}

func (mr *MentorshipRepository) GetRequestsByMentee(ctx context.Context, menteeID string, limit int, offset int) ([]mentorshippkg.MentorshipRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(menteeID)
	if err != nil {
		return nil, errors.New("invalid mentee ID")
	}

	filter := bson.M{"menteeId": objectID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"createdAt": -1})

	cursor, err := mr.requestsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []mentorshippkg.MentorshipRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (mr *MentorshipRepository) GetRequestsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]mentorshippkg.MentorshipRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(mentorID)
	if err != nil {
		return nil, errors.New("invalid mentor ID")
	}

	filter := bson.M{"mentorId": objectID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"createdAt": -1})

	cursor, err := mr.requestsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []mentorshippkg.MentorshipRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (mr *MentorshipRepository) GetPendingRequestsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]mentorshippkg.MentorshipRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(mentorID)
	if err != nil {
		return nil, errors.New("invalid mentor ID")
	}

	filter := bson.M{
		"mentorId": objectID,
		"status":   mentorshippkg.StatusPending,
	}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"createdAt": -1})

	cursor, err := mr.requestsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []mentorshippkg.MentorshipRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (mr *MentorshipRepository) UpdateRequestStatus(ctx context.Context, requestID string, status mentorshippkg.MentorshipRequestStatus) error {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("invalid request ID")
	}

	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"updatedAt":   time.Now(),
			"responsedAt": time.Now(),
		},
	}

	result, err := mr.requestsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mentorshippkg.ErrRequestNotFound
	}

	return nil
}

func (mr *MentorshipRepository) DeleteRequest(ctx context.Context, requestID string) error {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return errors.New("invalid request ID")
	}

	result, err := mr.requestsCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mentorshippkg.ErrRequestNotFound
	}

	return nil
}

func (mr *MentorshipRepository) ExistsPendingRequest(ctx context.Context, menteeID, mentorID string) (bool, error) {
	menteeObjectID, err := primitive.ObjectIDFromHex(menteeID)
	if err != nil {
		return false, errors.New("invalid mentee ID")
	}

	mentorObjectID, err := primitive.ObjectIDFromHex(mentorID)
	if err != nil {
		return false, errors.New("invalid mentor ID")
	}

	filter := bson.M{
		"menteeId": menteeObjectID,
		"mentorId": mentorObjectID,
		"status":   mentorshippkg.StatusPending,
	}

	count, err := mr.requestsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Mentorship Connection operations

func (mr *MentorshipRepository) CreateConnection(ctx context.Context, connection mentorshippkg.MentorshipConnection) (mentorshippkg.MentorshipConnection, error) {
	connection.ID = primitive.NewObjectID()
	connection.CreatedAt = time.Now()
	connection.UpdatedAt = time.Now()
	connection.StartedAt = time.Now()
	connection.Status = mentorshippkg.ConnectionActive

	_, err := mr.connectionsCollection.InsertOne(ctx, connection)
	if err != nil {
		return mentorshippkg.MentorshipConnection{}, err
	}

	return connection, nil
}

func (mr *MentorshipRepository) GetConnectionByID(ctx context.Context, connectionID string) (mentorshippkg.MentorshipConnection, error) {
	objectID, err := primitive.ObjectIDFromHex(connectionID)
	if err != nil {
		return mentorshippkg.MentorshipConnection{}, errors.New("invalid connection ID")
	}

	var connection mentorshippkg.MentorshipConnection
	err = mr.connectionsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&connection)
	if err != nil {
		return mentorshippkg.MentorshipConnection{}, err
	}

	return connection, nil
}

func (mr *MentorshipRepository) GetConnectionsByMentee(ctx context.Context, menteeID string, limit int, offset int) ([]mentorshippkg.MentorshipConnection, error) {
	objectID, err := primitive.ObjectIDFromHex(menteeID)
	if err != nil {
		return nil, errors.New("invalid mentee ID")
	}

	filter := bson.M{"menteeId": objectID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"startedAt": -1})

	cursor, err := mr.connectionsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []mentorshippkg.MentorshipConnection
	if err = cursor.All(ctx, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

func (mr *MentorshipRepository) GetConnectionsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]mentorshippkg.MentorshipConnection, error) {
	objectID, err := primitive.ObjectIDFromHex(mentorID)
	if err != nil {
		return nil, errors.New("invalid mentor ID")
	}

	filter := bson.M{"mentorId": objectID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"startedAt": -1})

	cursor, err := mr.connectionsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []mentorshippkg.MentorshipConnection
	if err = cursor.All(ctx, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

func (mr *MentorshipRepository) GetActiveConnectionsByUser(ctx context.Context, userID string) ([]mentorshippkg.MentorshipConnection, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	filter := bson.M{
		"$or": []bson.M{
			{"menteeId": objectID},
			{"mentorId": objectID},
		},
		"status": mentorshippkg.ConnectionActive,
	}

	cursor, err := mr.connectionsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []mentorshippkg.MentorshipConnection
	if err = cursor.All(ctx, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

func (mr *MentorshipRepository) UpdateConnectionStatus(ctx context.Context, connectionID string, status mentorshippkg.MentorshipConnectionStatus) error {
	objectID, err := primitive.ObjectIDFromHex(connectionID)
	if err != nil {
		return errors.New("invalid connection ID")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	result, err := mr.connectionsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mentorshippkg.ErrConnectionNotFound
	}

	return nil
}

func (mr *MentorshipRepository) UpdateLastInteraction(ctx context.Context, connectionID string) error {
	objectID, err := primitive.ObjectIDFromHex(connectionID)
	if err != nil {
		return errors.New("invalid connection ID")
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"lastInteraction": &now,
			"updatedAt":       now,
		},
	}

	result, err := mr.connectionsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mentorshippkg.ErrConnectionNotFound
	}

	return nil
}

func (mr *MentorshipRepository) EndConnection(ctx context.Context, connectionID string, endReason string, rating *int, feedback string, endedByMentor bool) error {
	objectID, err := primitive.ObjectIDFromHex(connectionID)
	if err != nil {
		return errors.New("invalid connection ID")
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":    mentorshippkg.ConnectionEnded,
			"endedAt":   &now,
			"endReason": endReason,
			"updatedAt": now,
		},
	}

	// Add rating and feedback based on who is ending the connection
	if rating != nil {
		if endedByMentor {
			update["$set"].(bson.M)["mentorRating"] = rating
			update["$set"].(bson.M)["mentorFeedback"] = feedback
		} else {
			update["$set"].(bson.M)["menteeRating"] = rating
			update["$set"].(bson.M)["menteeFeedback"] = feedback
		}
	}

	result, err := mr.connectionsCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mentorshippkg.ErrConnectionNotFound
	}

	return nil
}

func (mr *MentorshipRepository) GetConnectionByRequestID(ctx context.Context, requestID string) (mentorshippkg.MentorshipConnection, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return mentorshippkg.MentorshipConnection{}, errors.New("invalid request ID")
	}

	var connection mentorshippkg.MentorshipConnection
	err = mr.connectionsCollection.FindOne(ctx, bson.M{"requestId": objectID}).Decode(&connection)
	if err != nil {
		return mentorshippkg.MentorshipConnection{}, err
	}

	return connection, nil
}

// Analytics and statistics
func (mr *MentorshipRepository) GetMentorshipStats(ctx context.Context, userID string) (mentorshippkg.MentorshipStats, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return mentorshippkg.MentorshipStats{}, errors.New("invalid user ID")
	}

	stats := mentorshippkg.MentorshipStats{
		UserID: objectID,
	}

	// Get mentor statistics
	mentorRequestsCount, _ := mr.requestsCollection.CountDocuments(ctx, bson.M{"mentorId": objectID})
	stats.TotalMentorRequests = int(mentorRequestsCount)

	acceptedMentorRequestsCount, _ := mr.requestsCollection.CountDocuments(ctx, bson.M{
		"mentorId": objectID,
		"status":   mentorshippkg.StatusAccepted,
	})
	stats.AcceptedMentorRequests = int(acceptedMentorRequestsCount)

	activeMentorshipsCount, _ := mr.connectionsCollection.CountDocuments(ctx, bson.M{
		"mentorId": objectID,
		"status":   mentorshippkg.ConnectionActive,
	})
	stats.ActiveMentorships = int(activeMentorshipsCount)

	completedMentorshipsCount, _ := mr.connectionsCollection.CountDocuments(ctx, bson.M{
		"mentorId": objectID,
		"status": bson.M{"$in": []mentorshippkg.MentorshipConnectionStatus{
			mentorshippkg.ConnectionCompleted,
			mentorshippkg.ConnectionEnded,
		}},
	})
	stats.CompletedMentorships = int(completedMentorshipsCount)

	// Get mentee statistics
	menteeRequestsCount, _ := mr.requestsCollection.CountDocuments(ctx, bson.M{"menteeId": objectID})
	stats.TotalMenteeRequests = int(menteeRequestsCount)

	acceptedMenteeRequestsCount, _ := mr.requestsCollection.CountDocuments(ctx, bson.M{
		"menteeId": objectID,
		"status":   mentorshippkg.StatusAccepted,
	})
	stats.AcceptedMenteeRequests = int(acceptedMenteeRequestsCount)

	activeMenteeshipsCount, _ := mr.connectionsCollection.CountDocuments(ctx, bson.M{
		"menteeId": objectID,
		"status":   mentorshippkg.ConnectionActive,
	})
	stats.ActiveMenteerships = int(activeMenteeshipsCount)

	completedMenteeshipsCount, _ := mr.connectionsCollection.CountDocuments(ctx, bson.M{
		"menteeId": objectID,
		"status": bson.M{"$in": []mentorshippkg.MentorshipConnectionStatus{
			mentorshippkg.ConnectionCompleted,
			mentorshippkg.ConnectionEnded,
		}},
	})
	stats.CompletedMenteerships = int(completedMenteeshipsCount)

	// Calculate total connections
	stats.TotalConnections = stats.CompletedMentorships + stats.CompletedMenteerships

	return stats, nil
}

// Search operations
func (mr *MentorshipRepository) SearchRequests(ctx context.Context, filters mentorshippkg.RequestFilters) ([]mentorshippkg.MentorshipRequest, error) {
	filter := bson.M{}

	if filters.MenteeID != nil {
		filter["menteeId"] = *filters.MenteeID
	}
	if filters.MentorID != nil {
		filter["mentorId"] = *filters.MentorID
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if len(filters.Topics) > 0 {
		filter["topics"] = bson.M{"$in": filters.Topics}
	}

	opts := options.Find().SetLimit(int64(filters.Limit)).SetSkip(int64(filters.Offset)).SetSort(bson.M{"createdAt": -1})

	cursor, err := mr.requestsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []mentorshippkg.MentorshipRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (mr *MentorshipRepository) SearchConnections(ctx context.Context, filters mentorshippkg.ConnectionFilters) ([]mentorshippkg.MentorshipConnection, error) {
	filter := bson.M{}

	if filters.MenteeID != nil {
		filter["menteeId"] = *filters.MenteeID
	}
	if filters.MentorID != nil {
		filter["mentorId"] = *filters.MentorID
	}
	if filters.Status != nil {
		filter["status"] = *filters.Status
	}
	if len(filters.Topics) > 0 {
		filter["topics"] = bson.M{"$in": filters.Topics}
	}

	opts := options.Find().SetLimit(int64(filters.Limit)).SetSkip(int64(filters.Offset)).SetSort(bson.M{"startedAt": -1})

	cursor, err := mr.connectionsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []mentorshippkg.MentorshipConnection
	if err = cursor.All(ctx, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}
