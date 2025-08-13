package mentorshippkg

import (
	"context"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IMentorshipUsecase defines mentorship business logic operations
type IMentorshipUsecase interface {
	// Mentorship Request operations
	SendMentorshipRequest(ctx context.Context, menteeID string, request CreateMentorshipRequestDTO) (MentorshipRequestResponse, error)
	GetMentorshipRequest(ctx context.Context, requestID string, userID string) (MentorshipRequestResponse, error)
	GetIncomingRequests(ctx context.Context, mentorID string, limit int, offset int) ([]MentorshipRequestResponse, error)
	GetOutgoingRequests(ctx context.Context, menteeID string, limit int, offset int) ([]MentorshipRequestResponse, error)
	RespondToRequest(ctx context.Context, requestID string, mentorID string, response RespondToRequestDTO) (MentorshipRequestResponse, error)
	CancelRequest(ctx context.Context, requestID string, userID string) error

	// Mentorship Connection operations
	GetMentorshipConnection(ctx context.Context, connectionID string, userID string) (MentorshipConnectionResponse, error)
	GetMyMentorships(ctx context.Context, userID string, limit int, offset int) ([]MentorshipConnectionResponse, error)
	GetMyMenteerships(ctx context.Context, userID string, limit int, offset int) ([]MentorshipConnectionResponse, error)
	GetActiveConnections(ctx context.Context, userID string) ([]MentorshipConnectionResponse, error)
	UpdateLastInteraction(ctx context.Context, connectionID string, userID string) error
	PauseConnection(ctx context.Context, connectionID string, userID string) error
	ResumeConnection(ctx context.Context, connectionID string, userID string) error
	EndConnection(ctx context.Context, connectionID string, userID string, endData EndConnectionDTO) error

	// Analytics and insights
	GetMentorshipStats(ctx context.Context, userID string) (MentorshipStats, error)
	GetMentorshipInsights(ctx context.Context, userID string) (MentorshipInsights, error)

	// Search and discovery
	SearchMentorshipRequests(ctx context.Context, userID string, filters RequestFilters) ([]MentorshipRequestResponse, error)
	SearchMentorshipConnections(ctx context.Context, userID string, filters ConnectionFilters) ([]MentorshipConnectionResponse, error)

	// Validation and business rules
	CanSendRequest(ctx context.Context, menteeID string, mentorID string) error
	ValidateRequestAccess(ctx context.Context, requestID string, userID string) error
	ValidateConnectionAccess(ctx context.Context, connectionID string, userID string) error
}

// Additional response types for insights and analytics
type MentorshipInsights struct {
	UserID primitive.ObjectID `json:"userId"`

	// Recent activity
	RecentRequests    []MentorshipRequestResponse    `json:"recentRequests"`
	RecentConnections []MentorshipConnectionResponse `json:"recentConnections"`

	// Recommendations
	SuggestedMentors []userpkg.PublicProfile `json:"suggestedMentors,omitempty"`
	SuggestedMentees []userpkg.PublicProfile `json:"suggestedMentees,omitempty"`

	// Performance metrics
	ResponseRate          *float64 `json:"responseRate,omitempty"`          // For mentors
	AverageResponseTime   *int     `json:"averageResponseTime,omitempty"`   // In hours
	ConnectionSuccessRate *float64 `json:"connectionSuccessRate,omitempty"` // Requests that become connections

	// Engagement metrics
	ActiveConnectionsCount int     `json:"activeConnectionsCount"`
	TotalInteractionsCount int     `json:"totalInteractionsCount"`
	LastActivityDate       *string `json:"lastActivityDate,omitempty"`

	// Topic insights
	MostRequestedTopics  []TopicCount `json:"mostRequestedTopics"`
	MostSuccessfulTopics []TopicCount `json:"mostSuccessfulTopics"`
}

// Business rule constants
const (
	MaxActiveConnections     = 5  // Maximum active connections per user
	MaxPendingRequests       = 10 // Maximum pending requests per mentee
	RequestExpirationDays    = 30 // Days after which pending requests expire
	ConnectionInactivityDays = 90 // Days of inactivity before connection is marked stale
)

// Notification types for mentorship events
type NotificationType string

const (
	NotificationRequestReceived   NotificationType = "request_received"
	NotificationRequestAccepted   NotificationType = "request_accepted"
	NotificationRequestRejected   NotificationType = "request_rejected"
	NotificationConnectionEnded   NotificationType = "connection_ended"
	NotificationConnectionPaused  NotificationType = "connection_paused"
	NotificationConnectionResumed NotificationType = "connection_resumed"
	NotificationRatingReceived    NotificationType = "rating_received"
)

// Notification payload for mentorship events
type MentorshipNotification struct {
	Type         NotificationType `json:"type"`
	RecipientID  string           `json:"recipientId"`
	SenderID     string           `json:"senderId"`
	RequestID    *string          `json:"requestId,omitempty"`
	ConnectionID *string          `json:"connectionId,omitempty"`
	Message      string           `json:"message"`
	Data         interface{}      `json:"data,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
}

// Service interfaces that mentorship usecase depends on
type INotificationService interface {
	SendMentorshipNotification(ctx context.Context, notification MentorshipNotification) error
}

type IUserService interface {
	GetPublicProfile(ctx context.Context, userID string) (userpkg.PublicProfile, error)
	GetUserInfo(ctx context.Context, userID string, includePrivateInfo bool) (UserInfo, error)
	IsUserAvailableForMentoring(ctx context.Context, userID string) (bool, error)
	UpdateMentorAvailability(ctx context.Context, userID string, available bool) error
}

// Business rule validation errors
var (
	ErrMaxActiveConnectionsReached = NewMentorshipError("maximum active connections reached")
	ErrMaxPendingRequestsReached   = NewMentorshipError("maximum pending requests reached")
	ErrRequestExpired              = NewMentorshipError("mentorship request has expired")
	ErrConnectionInactive          = NewMentorshipError("connection has been inactive for too long")
	ErrInvalidStatusTransition     = NewMentorshipError("invalid status transition")
	ErrAlreadyRated                = NewMentorshipError("connection has already been rated")
	ErrCannotRateActiveConnection  = NewMentorshipError("cannot rate an active connection")
)

// Helper functions for business rule validation
func IsValidTopicList(topics []string, availableTopics []string) bool {
	if len(topics) == 0 {
		return false
	}

	topicMap := make(map[string]bool)
	for _, topic := range availableTopics {
		topicMap[topic] = true
	}

	for _, topic := range topics {
		if !topicMap[topic] {
			return false
		}
	}

	return true
}

func CalculateResponseRate(totalRequests, respondedRequests int) float64 {
	if totalRequests == 0 {
		return 0.0
	}
	return float64(respondedRequests) / float64(totalRequests) * 100.0
}

func CalculateSuccessRate(totalRequests, acceptedRequests int) float64 {
	if totalRequests == 0 {
		return 0.0
	}
	return float64(acceptedRequests) / float64(totalRequests) * 100.0
}

// Privacy-aware user info builder
func BuildUserInfoFromProfile(profile userpkg.PublicProfile, includePrivateInfo bool) UserInfo {
	userInfo := UserInfo{
		ID:                    profile.ID,
		DisplayName:           profile.DisplayName,
		Bio:                   profile.Bio,
		ProfilePicture:        profile.ProfilePicture,
		MentorshipBio:         profile.MentorshipBio,
		MentorshipTopics:      profile.MentorshipTopics,
		AvailableForMentoring: profile.AvailableForMentoring,
	}

	if includePrivateInfo {
		userInfo.Fullname = profile.Fullname
		userInfo.ContactInfo.Phone = profile.ContactInfo.Phone
		userInfo.ContactInfo.Website = profile.ContactInfo.Website
		userInfo.ContactInfo.Twitter = profile.ContactInfo.Twitter
		userInfo.ContactInfo.LinkedIn = profile.ContactInfo.LinkedIn
	}

	return userInfo
}
