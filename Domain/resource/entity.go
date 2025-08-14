package resourcepkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Resource represents a guide, tip, or opportunity shared on the platform
type Resource struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatorID   primitive.ObjectID `bson:"creatorId" json:"creatorId"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Content     string             `bson:"content" json:"content"`
	Type        string             `bson:"type" json:"type"` // "guide", "tip", "opportunity", "tool"
	Category    string             `bson:"category" json:"category"`
	Tags        []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	
	// Resource-specific fields
	ExternalURL    string      `bson:"externalUrl,omitempty" json:"externalUrl,omitempty"`
	Attachments    []Attachment `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Difficulty     string      `bson:"difficulty,omitempty" json:"difficulty,omitempty"` // "beginner", "intermediate", "advanced"
	EstimatedTime  string      `bson:"estimatedTime,omitempty" json:"estimatedTime,omitempty"` // e.g., "30 minutes", "2 hours"
	Prerequisites  []string    `bson:"prerequisites,omitempty" json:"prerequisites,omitempty"`
	
	// Opportunity-specific fields (for scholarships, internships, etc.)
	Deadline       *time.Time `bson:"deadline,omitempty" json:"deadline,omitempty"`
	ApplicationURL string     `bson:"applicationUrl,omitempty" json:"applicationUrl,omitempty"`
	Eligibility    []string   `bson:"eligibility,omitempty" json:"eligibility,omitempty"`
	Amount         string     `bson:"amount,omitempty" json:"amount,omitempty"` // For scholarships/grants
	
	// Engagement metrics
	ViewsCount     int                  `bson:"viewsCount" json:"viewsCount"`
	LikesCount     int                  `bson:"likesCount" json:"likesCount"`
	BookmarksCount int                  `bson:"bookmarksCount" json:"bookmarksCount"`
	SharesCount    int                  `bson:"sharesCount" json:"sharesCount"`
	LikedBy        []primitive.ObjectID `bson:"likedBy,omitempty" json:"likedBy,omitempty"`
	BookmarkedBy   []primitive.ObjectID `bson:"bookmarkedBy,omitempty" json:"bookmarkedBy,omitempty"`
	
	// Quality and verification
	IsVerified     bool    `bson:"isVerified" json:"isVerified"`
	VerifiedBy     primitive.ObjectID `bson:"verifiedBy,omitempty" json:"verifiedBy,omitempty"`
	QualityScore   float64 `bson:"qualityScore" json:"qualityScore"`
	Rating         float64 `bson:"rating" json:"rating"`
	RatingCount    int     `bson:"ratingCount" json:"ratingCount"`
	
	// Metadata
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	
	// Moderation
	IsReported bool   `bson:"isReported" json:"isReported"`
	IsHidden   bool   `bson:"isHidden" json:"isHidden"`
	Status     string `bson:"status" json:"status"` // "active", "hidden", "deleted", "pending"
}

// Attachment represents a file or link attached to a resource
type Attachment struct {
	Type        string `bson:"type" json:"type"` // "file", "link", "video", "document"
	URL         string `bson:"url" json:"url"`
	Name        string `bson:"name" json:"name"`
	Size        int64  `bson:"size,omitempty" json:"size,omitempty"` // in bytes
	MimeType    string `bson:"mimeType,omitempty" json:"mimeType,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
}

// CreateResourceRequest represents the request to create a new resource
type CreateResourceRequest struct {
	Title          string       `json:"title" validate:"required,min=3,max=200"`
	Description    string       `json:"description" validate:"required,min=10,max=500"`
	Content        string       `json:"content" validate:"required,min=20"`
	Type           string       `json:"type" validate:"required"`
	Category       string       `json:"category" validate:"required"`
	Tags           []string     `json:"tags,omitempty"`
	ExternalURL    string       `json:"externalUrl,omitempty" validate:"omitempty,url"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	Difficulty     string       `json:"difficulty,omitempty"`
	EstimatedTime  string       `json:"estimatedTime,omitempty"`
	Prerequisites  []string     `json:"prerequisites,omitempty"`
	Deadline       *time.Time   `json:"deadline,omitempty"`
	ApplicationURL string       `json:"applicationUrl,omitempty" validate:"omitempty,url"`
	Eligibility    []string     `json:"eligibility,omitempty"`
	Amount         string       `json:"amount,omitempty"`
}

// UpdateResourceRequest represents the request to update an existing resource
type UpdateResourceRequest struct {
	Title          string       `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description    string       `json:"description,omitempty" validate:"omitempty,min=10,max=500"`
	Content        string       `json:"content,omitempty" validate:"omitempty,min=20"`
	Type           string       `json:"type,omitempty"`
	Category       string       `json:"category,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	ExternalURL    string       `json:"externalUrl,omitempty" validate:"omitempty,url"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	Difficulty     string       `json:"difficulty,omitempty"`
	EstimatedTime  string       `json:"estimatedTime,omitempty"`
	Prerequisites  []string     `json:"prerequisites,omitempty"`
	Deadline       *time.Time   `json:"deadline,omitempty"`
	ApplicationURL string       `json:"applicationUrl,omitempty" validate:"omitempty,url"`
	Eligibility    []string     `json:"eligibility,omitempty"`
	Amount         string       `json:"amount,omitempty"`
}

// ResourceResponse represents a resource with creator information for API responses
type ResourceResponse struct {
	ID             primitive.ObjectID `json:"id"`
	Creator        CreatorInfo        `json:"creator"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	Content        string             `json:"content"`
	Type           string             `json:"type"`
	Category       string             `json:"category"`
	Tags           []string           `json:"tags,omitempty"`
	ExternalURL    string             `json:"externalUrl,omitempty"`
	Attachments    []Attachment       `json:"attachments,omitempty"`
	Difficulty     string             `json:"difficulty,omitempty"`
	EstimatedTime  string             `json:"estimatedTime,omitempty"`
	Prerequisites  []string           `json:"prerequisites,omitempty"`
	Deadline       *time.Time         `json:"deadline,omitempty"`
	ApplicationURL string             `json:"applicationUrl,omitempty"`
	Eligibility    []string           `json:"eligibility,omitempty"`
	Amount         string             `json:"amount,omitempty"`
	
	// Engagement metrics
	ViewsCount      int     `json:"viewsCount"`
	LikesCount      int     `json:"likesCount"`
	BookmarksCount  int     `json:"bookmarksCount"`
	SharesCount     int     `json:"sharesCount"`
	IsLikedByUser   bool    `json:"isLikedByUser,omitempty"`
	IsBookmarkedByUser bool `json:"isBookmarkedByUser,omitempty"`
	
	// Quality metrics
	IsVerified   bool    `json:"isVerified"`
	QualityScore float64 `json:"qualityScore"`
	Rating       float64 `json:"rating"`
	RatingCount  int     `json:"ratingCount"`
	
	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreatorInfo represents creator information in resource responses
type CreatorInfo struct {
	ID             primitive.ObjectID `json:"id"`
	DisplayName    string             `json:"displayName"`
	ProfilePicture string             `json:"profilePicture,omitempty"`
	IsMentor       bool               `json:"isMentor"`
	IsVerified     bool               `json:"isVerified"`
}

// ResourceFilter represents filtering options for resources
type ResourceFilter struct {
	Type        string   `json:"type,omitempty"`
	Category    string   `json:"category,omitempty"`
	CreatorID   string   `json:"creatorId,omitempty"`
	Tag         string   `json:"tag,omitempty"`
	Difficulty  string   `json:"difficulty,omitempty"`
	IsVerified  *bool    `json:"isVerified,omitempty"`
	HasDeadline *bool    `json:"hasDeadline,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ResourcePagination represents pagination options for resources
type ResourcePagination struct {
	Page      int    `json:"page" validate:"min=1"`
	PageSize  int    `json:"pageSize" validate:"min=1,max=100"`
	SortBy    string `json:"sortBy,omitempty"` // "createdAt", "rating", "viewsCount", "deadline"
	SortOrder string `json:"sortOrder,omitempty"` // "asc", "desc"
}

// ResourceTypes defines available resource types
var ResourceTypes = []string{
	"guide",
	"tip", 
	"opportunity",
	"tool",
	"template",
	"checklist",
	"tutorial",
}

// ResourceCategories defines available resource categories for ShareSpace
var ResourceCategories = []string{
	"Academic Success",
	"Financial Aid",
	"Scholarships",
	"Internships",
	"Career Development",
	"Mental Health Resources",
	"Study Tools",
	"Time Management",
	"Research Methods",
	"Writing Skills",
	"Technology Skills",
	"Life Skills",
	"Campus Resources",
	"Emergency Support",
	"Health & Wellness",
}

// DifficultyLevels defines available difficulty levels
var DifficultyLevels = []string{
	"beginner",
	"intermediate", 
	"advanced",
}

// ResourceStatus constants
const (
	ResourceStatusActive  = "active"
	ResourceStatusHidden  = "hidden"
	ResourceStatusDeleted = "deleted"
	ResourceStatusPending = "pending"
)

// AttachmentType constants
const (
	AttachmentTypeFile     = "file"
	AttachmentTypeLink     = "link"
	AttachmentTypeVideo    = "video"
	AttachmentTypeDocument = "document"
)
