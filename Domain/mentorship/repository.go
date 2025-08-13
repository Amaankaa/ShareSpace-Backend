package mentorshippkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IMentorshipRepository defines mentorship data access operations
type IMentorshipRepository interface {
	// Mentorship Request operations
	CreateRequest(ctx context.Context, request MentorshipRequest) (MentorshipRequest, error)
	GetRequestByID(ctx context.Context, requestID string) (MentorshipRequest, error)
	GetRequestsByMentee(ctx context.Context, menteeID string, limit int, offset int) ([]MentorshipRequest, error)
	GetRequestsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]MentorshipRequest, error)
	GetPendingRequestsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]MentorshipRequest, error)
	UpdateRequestStatus(ctx context.Context, requestID string, status MentorshipRequestStatus) error
	DeleteRequest(ctx context.Context, requestID string) error
	ExistsPendingRequest(ctx context.Context, menteeID, mentorID string) (bool, error)
	
	// Mentorship Connection operations
	CreateConnection(ctx context.Context, connection MentorshipConnection) (MentorshipConnection, error)
	GetConnectionByID(ctx context.Context, connectionID string) (MentorshipConnection, error)
	GetConnectionsByMentee(ctx context.Context, menteeID string, limit int, offset int) ([]MentorshipConnection, error)
	GetConnectionsByMentor(ctx context.Context, mentorID string, limit int, offset int) ([]MentorshipConnection, error)
	GetActiveConnectionsByUser(ctx context.Context, userID string) ([]MentorshipConnection, error)
	UpdateConnectionStatus(ctx context.Context, connectionID string, status MentorshipConnectionStatus) error
	UpdateLastInteraction(ctx context.Context, connectionID string) error
	EndConnection(ctx context.Context, connectionID string, endReason string, rating *int, feedback string, endedByMentor bool) error
	
	// Analytics and statistics
	GetMentorshipStats(ctx context.Context, userID string) (MentorshipStats, error)
	GetConnectionByRequestID(ctx context.Context, requestID string) (MentorshipConnection, error)
	
	// Search and filtering
	SearchRequests(ctx context.Context, filters RequestFilters) ([]MentorshipRequest, error)
	SearchConnections(ctx context.Context, filters ConnectionFilters) ([]MentorshipConnection, error)
}

// Filter structs for search operations
type RequestFilters struct {
	MenteeID   *primitive.ObjectID      `json:"menteeId,omitempty"`
	MentorID   *primitive.ObjectID      `json:"mentorId,omitempty"`
	Status     *MentorshipRequestStatus `json:"status,omitempty"`
	Topics     []string                 `json:"topics,omitempty"`
	DateFrom   *string                  `json:"dateFrom,omitempty"`   // ISO date string
	DateTo     *string                  `json:"dateTo,omitempty"`     // ISO date string
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
}

type ConnectionFilters struct {
	MenteeID   *primitive.ObjectID          `json:"menteeId,omitempty"`
	MentorID   *primitive.ObjectID          `json:"mentorId,omitempty"`
	Status     *MentorshipConnectionStatus  `json:"status,omitempty"`
	Topics     []string                     `json:"topics,omitempty"`
	DateFrom   *string                      `json:"dateFrom,omitempty"`   // ISO date string
	DateTo     *string                      `json:"dateTo,omitempty"`     // ISO date string
	Limit      int                          `json:"limit"`
	Offset     int                          `json:"offset"`
}

// Statistics struct for mentorship analytics
type MentorshipStats struct {
	UserID                primitive.ObjectID `json:"userId"`
	
	// As Mentor
	TotalMentorRequests   int `json:"totalMentorRequests"`
	AcceptedMentorRequests int `json:"acceptedMentorRequests"`
	ActiveMentorships     int `json:"activeMentorships"`
	CompletedMentorships  int `json:"completedMentorships"`
	AverageRatingAsMentor *float64 `json:"averageRatingAsMentor,omitempty"`
	
	// As Mentee
	TotalMenteeRequests   int `json:"totalMenteeRequests"`
	AcceptedMenteeRequests int `json:"acceptedMenteeRequests"`
	ActiveMenteerships    int `json:"activeMenteerships"`
	CompletedMenteerships int `json:"completedMenteerships"`
	AverageRatingAsMentee *float64 `json:"averageRatingAsMentee,omitempty"`
	
	// General
	TotalConnections      int `json:"totalConnections"`
	MostPopularTopics     []TopicCount `json:"mostPopularTopics"`
}

type TopicCount struct {
	Topic string `json:"topic"`
	Count int    `json:"count"`
}

// Validation helpers for filters
func (f *RequestFilters) Validate() error {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20 // Default limit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return nil
}

func (f *ConnectionFilters) Validate() error {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20 // Default limit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return nil
}

// Helper functions for status validation
func IsValidRequestStatus(status MentorshipRequestStatus) bool {
	switch status {
	case StatusPending, StatusAccepted, StatusRejected, StatusCanceled:
		return true
	default:
		return false
	}
}

func IsValidConnectionStatus(status MentorshipConnectionStatus) bool {
	switch status {
	case ConnectionActive, ConnectionPaused, ConnectionCompleted, ConnectionEnded:
		return true
	default:
		return false
	}
}

// Status transition validation
func CanTransitionRequestStatus(from, to MentorshipRequestStatus) bool {
	switch from {
	case StatusPending:
		return to == StatusAccepted || to == StatusRejected || to == StatusCanceled
	case StatusAccepted:
		return to == StatusCanceled // Can cancel after acceptance but before connection
	case StatusRejected, StatusCanceled:
		return false // Terminal states
	default:
		return false
	}
}

func CanTransitionConnectionStatus(from, to MentorshipConnectionStatus) bool {
	switch from {
	case ConnectionActive:
		return to == ConnectionPaused || to == ConnectionCompleted || to == ConnectionEnded
	case ConnectionPaused:
		return to == ConnectionActive || to == ConnectionCompleted || to == ConnectionEnded
	case ConnectionCompleted, ConnectionEnded:
		return false // Terminal states
	default:
		return false
	}
}
