package usecases

import (
	"context"
	"errors"

	mentorshippkg "github.com/Amaankaa/Blog-Starter-Project/Domain/mentorship"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MentorshipUsecase struct {
	mentorshipRepo mentorshippkg.IMentorshipRepository
	userRepo       userpkg.IUserRepository
}

func NewMentorshipUsecase(
	mentorshipRepo mentorshippkg.IMentorshipRepository,
	userRepo userpkg.IUserRepository,
) *MentorshipUsecase {
	return &MentorshipUsecase{
		mentorshipRepo: mentorshipRepo,
		userRepo:       userRepo,
	}
}

// SendMentorshipRequest creates a new mentorship request
func (mu *MentorshipUsecase) SendMentorshipRequest(ctx context.Context, menteeID string, request mentorshippkg.CreateMentorshipRequestDTO) (mentorshippkg.MentorshipRequestResponse, error) {
	// Validate the request
	if err := request.Validate(); err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Check if user can send request
	if err := mu.CanSendRequest(ctx, menteeID, request.MentorID.Hex()); err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Convert menteeID to ObjectID
	menteeObjectID, err := primitive.ObjectIDFromHex(menteeID)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, errors.New("invalid mentee ID")
	}

	// Create the mentorship request
	mentorshipRequest := mentorshippkg.MentorshipRequest{
		MenteeID: menteeObjectID,
		MentorID: request.MentorID,
		Message:  request.Message,
		Topics:   request.Topics,
	}

	createdRequest, err := mu.mentorshipRepo.CreateRequest(ctx, mentorshipRequest)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Build response with user information
	return mu.buildRequestResponse(ctx, createdRequest)
}

// GetMentorshipRequest retrieves a specific mentorship request
func (mu *MentorshipUsecase) GetMentorshipRequest(ctx context.Context, requestID string, userID string) (mentorshippkg.MentorshipRequestResponse, error) {
	// Validate access
	if err := mu.ValidateRequestAccess(ctx, requestID, userID); err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	request, err := mu.mentorshipRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	return mu.buildRequestResponse(ctx, request)
}

// GetIncomingRequests gets mentorship requests received by a mentor
func (mu *MentorshipUsecase) GetIncomingRequests(ctx context.Context, mentorID string, limit int, offset int) ([]mentorshippkg.MentorshipRequestResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	requests, err := mu.mentorshipRepo.GetPendingRequestsByMentor(ctx, mentorID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipRequestResponse
	for _, request := range requests {
		response, err := mu.buildRequestResponse(ctx, request)
		if err != nil {
			continue // Skip requests with errors
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetOutgoingRequests gets mentorship requests sent by a mentee
func (mu *MentorshipUsecase) GetOutgoingRequests(ctx context.Context, menteeID string, limit int, offset int) ([]mentorshippkg.MentorshipRequestResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	requests, err := mu.mentorshipRepo.GetRequestsByMentee(ctx, menteeID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipRequestResponse
	for _, request := range requests {
		response, err := mu.buildRequestResponse(ctx, request)
		if err != nil {
			continue // Skip requests with errors
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// RespondToRequest allows a mentor to accept or reject a mentorship request
func (mu *MentorshipUsecase) RespondToRequest(ctx context.Context, requestID string, mentorID string, response mentorshippkg.RespondToRequestDTO) (mentorshippkg.MentorshipRequestResponse, error) {
	// Validate access
	if err := mu.ValidateRequestAccess(ctx, requestID, mentorID); err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Get the request
	request, err := mu.mentorshipRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Verify the mentor is the intended recipient
	if request.MentorID.Hex() != mentorID {
		return mentorshippkg.MentorshipRequestResponse{}, mentorshippkg.ErrUnauthorizedAction
	}

	// Check if request is still pending
	if request.Status != mentorshippkg.StatusPending {
		return mentorshippkg.MentorshipRequestResponse{}, errors.New("request is no longer pending")
	}

	var newStatus mentorshippkg.MentorshipRequestStatus
	if response.Accept {
		newStatus = mentorshippkg.StatusAccepted

		// Create mentorship connection
		connection := mentorshippkg.MentorshipConnection{
			MenteeID:  request.MenteeID,
			MentorID:  request.MentorID,
			RequestID: request.ID,
			Topics:    request.Topics,
		}

		_, err := mu.mentorshipRepo.CreateConnection(ctx, connection)
		if err != nil {
			return mentorshippkg.MentorshipRequestResponse{}, err
		}
	} else {
		newStatus = mentorshippkg.StatusRejected
	}

	// Update request status
	err = mu.mentorshipRepo.UpdateRequestStatus(ctx, requestID, newStatus)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Get updated request
	updatedRequest, err := mu.mentorshipRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	return mu.buildRequestResponse(ctx, updatedRequest)
}

// CancelRequest allows a mentee to cancel their mentorship request
func (mu *MentorshipUsecase) CancelRequest(ctx context.Context, requestID string, userID string) error {
	// Validate access
	if err := mu.ValidateRequestAccess(ctx, requestID, userID); err != nil {
		return err
	}

	// Get the request
	request, err := mu.mentorshipRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Verify the user is the mentee
	if request.MenteeID.Hex() != userID {
		return mentorshippkg.ErrUnauthorizedAction
	}

	// Check if request can be cancelled
	if request.Status != mentorshippkg.StatusPending {
		return errors.New("only pending requests can be cancelled")
	}

	// Update status to cancelled
	return mu.mentorshipRepo.UpdateRequestStatus(ctx, requestID, mentorshippkg.StatusCanceled)
}

// GetMentorshipConnection retrieves a specific mentorship connection
func (mu *MentorshipUsecase) GetMentorshipConnection(ctx context.Context, connectionID string, userID string) (mentorshippkg.MentorshipConnectionResponse, error) {
	// Validate access
	if err := mu.ValidateConnectionAccess(ctx, connectionID, userID); err != nil {
		return mentorshippkg.MentorshipConnectionResponse{}, err
	}

	connection, err := mu.mentorshipRepo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return mentorshippkg.MentorshipConnectionResponse{}, err
	}

	return mu.buildConnectionResponse(ctx, connection, userID)
}

// GetMyMentorships gets connections where user is a mentor
func (mu *MentorshipUsecase) GetMyMentorships(ctx context.Context, userID string, limit int, offset int) ([]mentorshippkg.MentorshipConnectionResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	connections, err := mu.mentorshipRepo.GetConnectionsByMentor(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipConnectionResponse
	for _, connection := range connections {
		response, err := mu.buildConnectionResponse(ctx, connection, userID)
		if err != nil {
			continue // Skip connections with errors
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetMyMenteerships gets connections where user is a mentee
func (mu *MentorshipUsecase) GetMyMenteerships(ctx context.Context, userID string, limit int, offset int) ([]mentorshippkg.MentorshipConnectionResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	connections, err := mu.mentorshipRepo.GetConnectionsByMentee(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipConnectionResponse
	for _, connection := range connections {
		response, err := mu.buildConnectionResponse(ctx, connection, userID)
		if err != nil {
			continue // Skip connections with errors
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetActiveConnections gets all active connections for a user
func (mu *MentorshipUsecase) GetActiveConnections(ctx context.Context, userID string) ([]mentorshippkg.MentorshipConnectionResponse, error) {
	connections, err := mu.mentorshipRepo.GetActiveConnectionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipConnectionResponse
	for _, connection := range connections {
		response, err := mu.buildConnectionResponse(ctx, connection, userID)
		if err != nil {
			continue // Skip connections with errors
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// UpdateLastInteraction updates the last interaction timestamp for a connection
func (mu *MentorshipUsecase) UpdateLastInteraction(ctx context.Context, connectionID string, userID string) error {
	// Validate access
	if err := mu.ValidateConnectionAccess(ctx, connectionID, userID); err != nil {
		return err
	}

	return mu.mentorshipRepo.UpdateLastInteraction(ctx, connectionID)
}

// PauseConnection pauses an active mentorship connection
func (mu *MentorshipUsecase) PauseConnection(ctx context.Context, connectionID string, userID string) error {
	// Validate access
	if err := mu.ValidateConnectionAccess(ctx, connectionID, userID); err != nil {
		return err
	}

	// Get connection to check current status
	connection, err := mu.mentorshipRepo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return err
	}

	if connection.Status != mentorshippkg.ConnectionActive {
		return errors.New("only active connections can be paused")
	}

	return mu.mentorshipRepo.UpdateConnectionStatus(ctx, connectionID, mentorshippkg.ConnectionPaused)
}

// ResumeConnection resumes a paused mentorship connection
func (mu *MentorshipUsecase) ResumeConnection(ctx context.Context, connectionID string, userID string) error {
	// Validate access
	if err := mu.ValidateConnectionAccess(ctx, connectionID, userID); err != nil {
		return err
	}

	// Get connection to check current status
	connection, err := mu.mentorshipRepo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return err
	}

	if connection.Status != mentorshippkg.ConnectionPaused {
		return errors.New("only paused connections can be resumed")
	}

	return mu.mentorshipRepo.UpdateConnectionStatus(ctx, connectionID, mentorshippkg.ConnectionActive)
}

// EndConnection ends a mentorship connection with optional rating and feedback
func (mu *MentorshipUsecase) EndConnection(ctx context.Context, connectionID string, userID string, endData mentorshippkg.EndConnectionDTO) error {
	// Validate the end data
	if err := endData.Validate(); err != nil {
		return err
	}

	// Validate access
	if err := mu.ValidateConnectionAccess(ctx, connectionID, userID); err != nil {
		return err
	}

	// Get connection to check current status and determine user role
	connection, err := mu.mentorshipRepo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return err
	}

	if connection.Status == mentorshippkg.ConnectionEnded || connection.Status == mentorshippkg.ConnectionCompleted {
		return errors.New("connection is already ended")
	}

	// Determine if user is mentor or mentee
	isMentor := connection.MentorID.Hex() == userID
	if !isMentor && connection.MenteeID.Hex() != userID {
		return mentorshippkg.ErrUnauthorizedAction
	}

	return mu.mentorshipRepo.EndConnection(ctx, connectionID, endData.Reason, endData.Rating, endData.Feedback, isMentor)
}

// GetMentorshipStats returns statistics for a user's mentorship activities
func (mu *MentorshipUsecase) GetMentorshipStats(ctx context.Context, userID string) (mentorshippkg.MentorshipStats, error) {
	return mu.mentorshipRepo.GetMentorshipStats(ctx, userID)
}

// GetMentorshipInsights returns insights and recommendations for a user
func (mu *MentorshipUsecase) GetMentorshipInsights(ctx context.Context, userID string) (mentorshippkg.MentorshipInsights, error) {
	// Get basic stats
	stats, err := mu.mentorshipRepo.GetMentorshipStats(ctx, userID)
	if err != nil {
		return mentorshippkg.MentorshipInsights{}, err
	}

	insights := mentorshippkg.MentorshipInsights{
		UserID:                 stats.UserID,
		ActiveConnectionsCount: stats.ActiveMentorships + stats.ActiveMenteerships,
	}

	// Calculate response rate for mentors
	if stats.TotalMentorRequests > 0 {
		responseRate := mentorshippkg.CalculateResponseRate(stats.TotalMentorRequests, stats.AcceptedMentorRequests)
		insights.ResponseRate = &responseRate
	}

	// Calculate connection success rate for mentees
	if stats.TotalMenteeRequests > 0 {
		successRate := mentorshippkg.CalculateSuccessRate(stats.TotalMenteeRequests, stats.AcceptedMenteeRequests)
		insights.ConnectionSuccessRate = &successRate
	}

	return insights, nil
}

// SearchMentorshipRequests searches mentorship requests with filters
func (mu *MentorshipUsecase) SearchMentorshipRequests(ctx context.Context, userID string, filters mentorshippkg.RequestFilters) ([]mentorshippkg.MentorshipRequestResponse, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	requests, err := mu.mentorshipRepo.SearchRequests(ctx, filters)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipRequestResponse
	for _, request := range requests {
		// Only include requests where user is involved
		if request.MenteeID.Hex() == userID || request.MentorID.Hex() == userID {
			response, err := mu.buildRequestResponse(ctx, request)
			if err != nil {
				continue
			}
			responses = append(responses, response)
		}
	}

	return responses, nil
}

// SearchMentorshipConnections searches mentorship connections with filters
func (mu *MentorshipUsecase) SearchMentorshipConnections(ctx context.Context, userID string, filters mentorshippkg.ConnectionFilters) ([]mentorshippkg.MentorshipConnectionResponse, error) {
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	connections, err := mu.mentorshipRepo.SearchConnections(ctx, filters)
	if err != nil {
		return nil, err
	}

	var responses []mentorshippkg.MentorshipConnectionResponse
	for _, connection := range connections {
		// Only include connections where user is involved
		if connection.MenteeID.Hex() == userID || connection.MentorID.Hex() == userID {
			response, err := mu.buildConnectionResponse(ctx, connection, userID)
			if err != nil {
				continue
			}
			responses = append(responses, response)
		}
	}

	return responses, nil
}

// Validation and business rule methods

// CanSendRequest validates if a mentee can send a request to a mentor
func (mu *MentorshipUsecase) CanSendRequest(ctx context.Context, menteeID string, mentorID string) error {
	// Check if trying to request themselves
	if menteeID == mentorID {
		return mentorshippkg.ErrCannotRequestSelf
	}

	// Check if mentor is available
	mentorProfile, err := mu.userRepo.GetPublicProfile(ctx, mentorID)
	if err != nil {
		return err
	}

	if !mentorProfile.IsMentor || !mentorProfile.AvailableForMentoring {
		return mentorshippkg.ErrMentorNotAvailable
	}

	// Check if pending request already exists
	exists, err := mu.mentorshipRepo.ExistsPendingRequest(ctx, menteeID, mentorID)
	if err != nil {
		return err
	}

	if exists {
		return mentorshippkg.ErrRequestAlreadyExists
	}

	// Check if mentee has reached max pending requests
	menteeRequests, err := mu.mentorshipRepo.GetRequestsByMentee(ctx, menteeID, mentorshippkg.MaxPendingRequests+1, 0)
	if err != nil {
		return err
	}

	pendingCount := 0
	for _, req := range menteeRequests {
		if req.Status == mentorshippkg.StatusPending {
			pendingCount++
		}
	}

	if pendingCount >= mentorshippkg.MaxPendingRequests {
		return mentorshippkg.ErrMaxPendingRequestsReached
	}

	// Check if mentor has reached max active connections
	activeConnections, err := mu.mentorshipRepo.GetActiveConnectionsByUser(ctx, mentorID)
	if err != nil {
		return err
	}

	if len(activeConnections) >= mentorshippkg.MaxActiveConnections {
		return mentorshippkg.ErrMaxActiveConnectionsReached
	}

	return nil
}

// ValidateRequestAccess checks if user has access to a mentorship request
func (mu *MentorshipUsecase) ValidateRequestAccess(ctx context.Context, requestID string, userID string) error {
	request, err := mu.mentorshipRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if request.MenteeID.Hex() != userID && request.MentorID.Hex() != userID {
		return mentorshippkg.ErrUnauthorizedAction
	}

	return nil
}

// ValidateConnectionAccess checks if user has access to a mentorship connection
func (mu *MentorshipUsecase) ValidateConnectionAccess(ctx context.Context, connectionID string, userID string) error {
	connection, err := mu.mentorshipRepo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return err
	}

	if connection.MenteeID.Hex() != userID && connection.MentorID.Hex() != userID {
		return mentorshippkg.ErrUnauthorizedAction
	}

	return nil
}

// Helper methods for building responses

// buildRequestResponse builds a MentorshipRequestResponse with user information
func (mu *MentorshipUsecase) buildRequestResponse(ctx context.Context, request mentorshippkg.MentorshipRequest) (mentorshippkg.MentorshipRequestResponse, error) {
	// Get mentee info
	menteeProfile, err := mu.userRepo.GetPublicProfile(ctx, request.MenteeID.Hex())
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	// Get mentor info
	mentorProfile, err := mu.userRepo.GetPublicProfile(ctx, request.MentorID.Hex())
	if err != nil {
		return mentorshippkg.MentorshipRequestResponse{}, err
	}

	return mentorshippkg.MentorshipRequestResponse{
		ID:          request.ID,
		MenteeInfo:  mentorshippkg.BuildUserInfoFromProfile(menteeProfile, false),
		MentorInfo:  mentorshippkg.BuildUserInfoFromProfile(mentorProfile, false),
		Status:      request.Status,
		Message:     request.Message,
		Topics:      request.Topics,
		CreatedAt:   request.CreatedAt,
		UpdatedAt:   request.UpdatedAt,
		ResponsedAt: request.ResponsedAt,
	}, nil
}

// buildConnectionResponse builds a MentorshipConnectionResponse with user information
func (mu *MentorshipUsecase) buildConnectionResponse(ctx context.Context, connection mentorshippkg.MentorshipConnection, currentUserID string) (mentorshippkg.MentorshipConnectionResponse, error) {
	// Get mentee info
	menteeProfile, err := mu.userRepo.GetPublicProfile(ctx, connection.MenteeID.Hex())
	if err != nil {
		return mentorshippkg.MentorshipConnectionResponse{}, err
	}

	// Get mentor info
	mentorProfile, err := mu.userRepo.GetPublicProfile(ctx, connection.MentorID.Hex())
	if err != nil {
		return mentorshippkg.MentorshipConnectionResponse{}, err
	}

	// Determine if current user can rate and has rated
	isMentor := connection.MentorID.Hex() == currentUserID
	canRate := connection.Status == mentorshippkg.ConnectionEnded || connection.Status == mentorshippkg.ConnectionCompleted
	hasRated := false

	if canRate {
		if isMentor {
			hasRated = connection.MentorRating != nil
		} else {
			hasRated = connection.MenteeRating != nil
		}
	}

	// Include private info for established connections
	includePrivateInfo := connection.Status == mentorshippkg.ConnectionActive ||
		connection.Status == mentorshippkg.ConnectionCompleted ||
		connection.Status == mentorshippkg.ConnectionEnded

	return mentorshippkg.MentorshipConnectionResponse{
		ID:              connection.ID,
		MenteeInfo:      mentorshippkg.BuildUserInfoFromProfile(menteeProfile, includePrivateInfo),
		MentorInfo:      mentorshippkg.BuildUserInfoFromProfile(mentorProfile, includePrivateInfo),
		Status:          connection.Status,
		Topics:          connection.Topics,
		StartedAt:       connection.StartedAt,
		LastInteraction: connection.LastInteraction,
		EndedAt:         connection.EndedAt,
		EndReason:       connection.EndReason,
		CanRate:         canRate,
		HasRated:        hasRated,
		CreatedAt:       connection.CreatedAt,
		UpdatedAt:       connection.UpdatedAt,
	}, nil
}
