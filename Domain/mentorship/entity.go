package mentorshippkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MentorshipRequestStatus represents the status of a mentorship request
type MentorshipRequestStatus string

const (
	StatusPending  MentorshipRequestStatus = "pending"
	StatusAccepted MentorshipRequestStatus = "accepted"
	StatusRejected MentorshipRequestStatus = "rejected"
	StatusCanceled MentorshipRequestStatus = "canceled"
)

// MentorshipRequest represents a request from a mentee to a mentor
type MentorshipRequest struct {
	ID          primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	MenteeID    primitive.ObjectID      `bson:"menteeId" json:"menteeId"`
	MentorID    primitive.ObjectID      `bson:"mentorId" json:"mentorId"`
	Status      MentorshipRequestStatus `bson:"status" json:"status"`
	Message     string                  `bson:"message,omitempty" json:"message,omitempty"`
	Topics      []string                `bson:"topics" json:"topics"`
	CreatedAt   time.Time               `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time               `bson:"updatedAt" json:"updatedAt"`
	ResponsedAt *time.Time              `bson:"responsedAt,omitempty" json:"responsedAt,omitempty"`
}

// MentorshipConnectionStatus represents the status of an active mentorship
type MentorshipConnectionStatus string

const (
	ConnectionActive    MentorshipConnectionStatus = "active"
	ConnectionPaused    MentorshipConnectionStatus = "paused"
	ConnectionCompleted MentorshipConnectionStatus = "completed"
	ConnectionEnded     MentorshipConnectionStatus = "ended"
)

// MentorshipConnection represents an active mentorship relationship
type MentorshipConnection struct {
	ID               primitive.ObjectID          `bson:"_id,omitempty" json:"id"`
	MenteeID         primitive.ObjectID          `bson:"menteeId" json:"menteeId"`
	MentorID         primitive.ObjectID          `bson:"mentorId" json:"mentorId"`
	RequestID        primitive.ObjectID          `bson:"requestId" json:"requestId"`
	Status           MentorshipConnectionStatus `bson:"status" json:"status"`
	Topics           []string                    `bson:"topics" json:"topics"`
	StartedAt        time.Time                   `bson:"startedAt" json:"startedAt"`
	LastInteraction  *time.Time                  `bson:"lastInteraction,omitempty" json:"lastInteraction,omitempty"`
	EndedAt          *time.Time                  `bson:"endedAt,omitempty" json:"endedAt,omitempty"`
	EndReason        string                      `bson:"endReason,omitempty" json:"endReason,omitempty"`
	MenteeRating     *int                        `bson:"menteeRating,omitempty" json:"menteeRating,omitempty"`     // 1-5 rating
	MentorRating     *int                        `bson:"mentorRating,omitempty" json:"mentorRating,omitempty"`     // 1-5 rating
	MenteeFeedback   string                      `bson:"menteeFeedback,omitempty" json:"menteeFeedback,omitempty"`
	MentorFeedback   string                      `bson:"mentorFeedback,omitempty" json:"mentorFeedback,omitempty"`
	CreatedAt        time.Time                   `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time                   `bson:"updatedAt" json:"updatedAt"`
}

// Request DTOs
type CreateMentorshipRequestDTO struct {
	MentorID primitive.ObjectID `json:"mentorId" binding:"required"`
	Message  string             `json:"message,omitempty"`
	Topics   []string           `json:"topics" binding:"required,min=1"`
}

type RespondToRequestDTO struct {
	Accept bool   `json:"accept" binding:"required"`
	Reason string `json:"reason,omitempty"` // Optional reason for rejection
}

type EndConnectionDTO struct {
	Reason         string `json:"reason,omitempty"`
	Rating         *int   `json:"rating,omitempty"`         // 1-5 rating
	Feedback       string `json:"feedback,omitempty"`
	EndedByMentor  bool   `json:"endedByMentor"`
}

// Response DTOs with privacy-aware user information
type MentorshipRequestResponse struct {
	ID          primitive.ObjectID      `json:"id"`
	MenteeInfo  UserInfo                `json:"menteeInfo"`
	MentorInfo  UserInfo                `json:"mentorInfo"`
	Status      MentorshipRequestStatus `json:"status"`
	Message     string                  `json:"message,omitempty"`
	Topics      []string                `json:"topics"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	ResponsedAt *time.Time              `json:"responsedAt,omitempty"`
}

type MentorshipConnectionResponse struct {
	ID               primitive.ObjectID          `json:"id"`
	MenteeInfo       UserInfo                    `json:"menteeInfo"`
	MentorInfo       UserInfo                    `json:"mentorInfo"`
	Status           MentorshipConnectionStatus `json:"status"`
	Topics           []string                    `json:"topics"`
	StartedAt        time.Time                   `json:"startedAt"`
	LastInteraction  *time.Time                  `json:"lastInteraction,omitempty"`
	EndedAt          *time.Time                  `json:"endedAt,omitempty"`
	EndReason        string                      `json:"endReason,omitempty"`
	CanRate          bool                        `json:"canRate"`          // Whether current user can rate
	HasRated         bool                        `json:"hasRated"`         // Whether current user has rated
	CreatedAt        time.Time                   `json:"createdAt"`
	UpdatedAt        time.Time                   `json:"updatedAt"`
}

// UserInfo represents privacy-aware user information for mentorship contexts
type UserInfo struct {
	ID                    primitive.ObjectID `json:"id"`
	DisplayName           string             `json:"displayName"`
	Bio                   string             `json:"bio,omitempty"`
	ProfilePicture        string             `json:"profilePicture,omitempty"`
	MentorshipBio         string             `json:"mentorshipBio,omitempty"`
	MentorshipTopics      []string           `json:"mentorshipTopics,omitempty"`
	AvailableForMentoring bool               `json:"availableForMentoring,omitempty"`
	
	// These fields are only included if privacy settings allow or in established connections
	Fullname    string `json:"fullname,omitempty"`
	ContactInfo struct {
		Phone    string `json:"phone,omitempty"`
		Website  string `json:"website,omitempty"`
		Twitter  string `json:"twitter,omitempty"`
		LinkedIn string `json:"linkedin,omitempty"`
	} `json:"contactInfo,omitempty"`
}

// Validation helpers
func (r *CreateMentorshipRequestDTO) Validate() error {
	if r.MentorID.IsZero() {
		return ErrInvalidMentorID
	}
	if len(r.Topics) == 0 {
		return ErrNoTopicsSpecified
	}
	if len(r.Message) > 500 {
		return ErrMessageTooLong
	}
	return nil
}

func (e *EndConnectionDTO) Validate() error {
	if e.Rating != nil && (*e.Rating < 1 || *e.Rating > 5) {
		return ErrInvalidRating
	}
	if len(e.Feedback) > 1000 {
		return ErrFeedbackTooLong
	}
	return nil
}

// Custom errors
var (
	ErrInvalidMentorID     = NewMentorshipError("invalid mentor ID")
	ErrNoTopicsSpecified   = NewMentorshipError("at least one topic must be specified")
	ErrMessageTooLong      = NewMentorshipError("message cannot exceed 500 characters")
	ErrInvalidRating       = NewMentorshipError("rating must be between 1 and 5")
	ErrFeedbackTooLong     = NewMentorshipError("feedback cannot exceed 1000 characters")
	ErrRequestNotFound     = NewMentorshipError("mentorship request not found")
	ErrConnectionNotFound  = NewMentorshipError("mentorship connection not found")
	ErrUnauthorizedAction  = NewMentorshipError("unauthorized to perform this action")
	ErrRequestAlreadyExists = NewMentorshipError("mentorship request already exists")
	ErrCannotRequestSelf   = NewMentorshipError("cannot send mentorship request to yourself")
	ErrMentorNotAvailable  = NewMentorshipError("mentor is not available for mentoring")
)

type MentorshipError struct {
	Message string
}

func (e MentorshipError) Error() string {
	return e.Message
}

func NewMentorshipError(message string) MentorshipError {
	return MentorshipError{Message: message}
}
