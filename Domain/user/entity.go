package userpkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user entity
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username       string             `bson:"username" json:"username"`
	Fullname       string             `bson:"fullname" json:"fullname"`
	Email          string             `bson:"email" json:"email"`
	Password       string             `bson:"password" json:"password"`
	Role           string             `bson:"role" json:"role"` // e.g. "admin", "user"
	IsVerified     bool               `bson:"isVerified" json:"isVerified"`
	Bio            string             `bson:"bio,omitempty" json:"bio,omitempty"`
	ProfilePicture string             `bson:"profilePicture,omitempty" json:"profilePicture,omitempty"`
	ContactInfo    ContactInfo        `bson:"contactInfo,omitempty" json:"contactInfo,omitempty"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	PromotedBy     primitive.ObjectID `bson:"promoted_by,omitempty" json:"promoted_by,omitempty"`

	// ShareSpace Anonymous Identity
	DisplayName string `bson:"displayName,omitempty" json:"displayName,omitempty"`
	IsAnonymous bool   `bson:"isAnonymous" json:"isAnonymous"`

	// ShareSpace Mentorship Features
	IsMentor              bool     `bson:"isMentor" json:"isMentor"`
	IsMentee              bool     `bson:"isMentee" json:"isMentee"`
	MentorshipTopics      []string `bson:"mentorshipTopics,omitempty" json:"mentorshipTopics,omitempty"`
	MentorshipBio         string   `bson:"mentorshipBio,omitempty" json:"mentorshipBio,omitempty"`
	AvailableForMentoring bool     `bson:"availableForMentoring" json:"availableForMentoring"`

	// Privacy Controls
	PrivacySettings PrivacySettings `bson:"privacySettings" json:"privacySettings"`
}

type ContactInfo struct {
	Phone    string `bson:"phone,omitempty" json:"phone,omitempty"`
	Website  string `bson:"website,omitempty" json:"website,omitempty"`
	Twitter  string `bson:"twitter,omitempty" json:"twitter,omitempty"`
	LinkedIn string `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
}

// PrivacySettings controls what information is visible to other users
type PrivacySettings struct {
	ShowRealName       bool `bson:"showRealName" json:"showRealName"`             // Default: false
	ShowProfilePicture bool `bson:"showProfilePicture" json:"showProfilePicture"` // Default: false
	ShowContactInfo    bool `bson:"showContactInfo" json:"showContactInfo"`       // Default: false
}

type UpdateProfileRequest struct {
	Fullname       string      `json:"fullname,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	ProfilePicture string      `json:"profilePicture,omitempty"`
	ContactInfo    ContactInfo `json:"contactInfo,omitempty"`

	// ShareSpace fields
	DisplayName           string          `json:"displayName,omitempty"`
	IsAnonymous           bool            `json:"isAnonymous,omitempty"`
	IsMentor              bool            `json:"isMentor,omitempty"`
	IsMentee              bool            `json:"isMentee,omitempty"`
	MentorshipTopics      []string        `json:"mentorshipTopics,omitempty"`
	MentorshipBio         string          `json:"mentorshipBio,omitempty"`
	AvailableForMentoring bool            `json:"availableForMentoring,omitempty"`
	PrivacySettings       PrivacySettings `json:"privacySettings,omitempty"`
}

// Token struct (We put it here since it's related with the User)
type Token struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	AccessToken  string             `bson:"access_token"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedAt    time.Time          `bson:"created_at"`
	ExpiresAt    time.Time          `bson:"expires_at"`
}

// Response upon login
type TokenResult struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type PasswordReset struct {
	Email        string    `bson:"email"`
	OTP          string    `bson:"otp"`
	ExpiresAt    time.Time `bson:"expiresat"`
	AttemptCount int       `bson:"attemptcount"`
}

type Verification struct {
	Email        string    `bson:"email" json:"email"`
	OTP          string    `bson:"otp" json:"otp"`
	ExpiresAt    time.Time `bson:"expiresAt" json:"expiresAt"`
	AttemptCount int       `bson:"attemptCount" json:"attemptCount"`
}

// PublicProfile represents what other users can see (respects privacy settings)
type PublicProfile struct {
	ID                    primitive.ObjectID `json:"id"`
	DisplayName           string             `json:"displayName"`
	Bio                   string             `json:"bio,omitempty"`
	ProfilePicture        string             `json:"profilePicture,omitempty"`
	IsMentor              bool               `json:"isMentor"`
	IsMentee              bool               `json:"isMentee"`
	MentorshipTopics      []string           `json:"mentorshipTopics,omitempty"`
	MentorshipBio         string             `json:"mentorshipBio,omitempty"`
	AvailableForMentoring bool               `json:"availableForMentoring"`

	// These fields are only included if privacy settings allow
	Fullname    string      `json:"fullname,omitempty"`
	ContactInfo ContactInfo `json:"contactInfo,omitempty"`
}

// MentorshipTopics defines available mentorship categories
var MentorshipTopics = []string{
	"Academic Support",
	"Career Guidance",
	"Study Techniques",
	"Time Management",
	"Mental Health & Wellness",
	"Technology Skills",
	"Research Methods",
	"Life Balance",
	"Communication Skills",
	"Leadership Development",
	"Networking",
	"Interview Preparation",
	"Project Management",
	"Creative Skills",
	"Personal Development",
}
