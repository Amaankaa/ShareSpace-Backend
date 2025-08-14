package postpkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a user's experience sharing post
type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthorID    primitive.ObjectID `bson:"authorId" json:"authorId"`
	Title       string             `bson:"title" json:"title"`
	Content     string             `bson:"content" json:"content"`
	Category    string             `bson:"category" json:"category"`
	Tags        []string           `bson:"tags,omitempty" json:"tags,omitempty"`
	MediaLinks  []MediaLink        `bson:"mediaLinks,omitempty" json:"mediaLinks,omitempty"`
	IsAnonymous bool               `bson:"isAnonymous" json:"isAnonymous"`
	
	// Engagement metrics
	LikesCount    int                  `bson:"likesCount" json:"likesCount"`
	CommentsCount int                  `bson:"commentsCount" json:"commentsCount"`
	ViewsCount    int                  `bson:"viewsCount" json:"viewsCount"`
	LikedBy       []primitive.ObjectID `bson:"likedBy,omitempty" json:"likedBy,omitempty"`
	
	// Metadata
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	
	// Moderation
	IsReported bool   `bson:"isReported" json:"isReported"`
	IsHidden   bool   `bson:"isHidden" json:"isHidden"`
	Status     string `bson:"status" json:"status"` // "active", "hidden", "deleted"
}

// MediaLink represents attached media in a post
type MediaLink struct {
	Type string `bson:"type" json:"type"` // "image", "video", "document", "link"
	URL  string `bson:"url" json:"url"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Title       string      `json:"title" validate:"required,min=3,max=200"`
	Content     string      `json:"content" validate:"required,min=10,max=5000"`
	Category    string      `json:"category" validate:"required"`
	Tags        []string    `json:"tags,omitempty"`
	MediaLinks  []MediaLink `json:"mediaLinks,omitempty"`
	IsAnonymous bool        `json:"isAnonymous"`
}

// UpdatePostRequest represents the request to update an existing post
type UpdatePostRequest struct {
	Title      string      `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Content    string      `json:"content,omitempty" validate:"omitempty,min=10,max=5000"`
	Category   string      `json:"category,omitempty"`
	Tags       []string    `json:"tags,omitempty"`
	MediaLinks []MediaLink `json:"mediaLinks,omitempty"`
}

// PostResponse represents a post with author information for API responses
type PostResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Author      AuthorInfo         `json:"author"`
	Title       string             `json:"title"`
	Content     string             `json:"content"`
	Category    string             `json:"category"`
	Tags        []string           `json:"tags,omitempty"`
	MediaLinks  []MediaLink        `json:"mediaLinks,omitempty"`
	IsAnonymous bool               `json:"isAnonymous"`
	
	// Engagement metrics
	LikesCount    int  `json:"likesCount"`
	CommentsCount int  `json:"commentsCount"`
	ViewsCount    int  `json:"viewsCount"`
	IsLikedByUser bool `json:"isLikedByUser,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AuthorInfo represents author information in post responses
type AuthorInfo struct {
	ID             primitive.ObjectID `json:"id"`
	DisplayName    string             `json:"displayName"`
	ProfilePicture string             `json:"profilePicture,omitempty"`
	IsMentor       bool               `json:"isMentor"`
	IsAnonymous    bool               `json:"isAnonymous"`
}

// PostFilter represents filtering options for posts
type PostFilter struct {
	Category   string `json:"category,omitempty"`
	AuthorID   string `json:"authorId,omitempty"`
	Tag        string `json:"tag,omitempty"`
	Year       int    `json:"year,omitempty"`
	Month      int    `json:"month,omitempty"`
	IsAnonymous *bool `json:"isAnonymous,omitempty"`
}

// PostPagination represents pagination options
type PostPagination struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=100"`
	SortBy   string `json:"sortBy,omitempty"` // "createdAt", "likesCount", "commentsCount", "viewsCount"
	SortOrder string `json:"sortOrder,omitempty"` // "asc", "desc"
}

// PostCategories defines available post categories for ShareSpace
var PostCategories = []string{
	"Academic Struggles",
	"Financial Challenges", 
	"Relationship Issues",
	"Mental Health",
	"Career Guidance",
	"Study Tips",
	"Campus Life",
	"Time Management",
	"Personal Growth",
	"Success Stories",
	"Failure Lessons",
	"Life Advice",
	"University Resources",
	"Extracurricular Activities",
	"General Discussion",
}

// PostStatus constants
const (
	PostStatusActive  = "active"
	PostStatusHidden  = "hidden"
	PostStatusDeleted = "deleted"
)

// MediaType constants
const (
	MediaTypeImage    = "image"
	MediaTypeVideo    = "video"
	MediaTypeDocument = "document"
	MediaTypeLink     = "link"
)
