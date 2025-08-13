package userpkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user entity
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username   string             `bson:"username" json:"username"`
	Fullname   string             `bson:"fullname" json:"fullname"`
	Email      string             `bson:"email" json:"email"`
	Password   string             `bson:"password" json:"password"`
	Role       string             `bson:"role" json:"role"` // e.g. "admin", "user"
	IsVerified bool               `bson:"isVerified" json:"isVerified"`
    Bio        string             `bson:"bio,omitempty" json:"bio,omitempty"`
    ProfilePicture  string         `bson:"profilePicture,omitempty" json:"profilePicture,omitempty"`
    ContactInfo ContactInfo        `bson:"contactInfo,omitempty" json:"contactInfo,omitempty"`
    UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	PromotedBy primitive.ObjectID `bson:"promoted_by,omitempty" json:"promoted_by,omitempty"`
}

type ContactInfo struct {
    Phone   string `bson:"phone,omitempty" json:"phone,omitempty"`
    Website string `bson:"website,omitempty" json:"website,omitempty"`
    Twitter string `bson:"twitter,omitempty" json:"twitter,omitempty"`
    LinkedIn string `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
}
type UpdateProfileRequest struct {
    Fullname        string      `json:"fullname,omitempty"`
    Bio             string      `json:"bio,omitempty"`
    ProfilePicture  string      `json:"profilePicture,omitempty"`
    ContactInfo     ContactInfo `json:"contactInfo,omitempty"`
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
