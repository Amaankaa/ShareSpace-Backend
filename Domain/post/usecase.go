package postpkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockery --name=PostUsecase --output=../../mocks --outpkg=mocks

// PostUsecase defines the business logic interface for posts
type PostUsecase interface {
	// Core post operations
	CreatePost(ctx context.Context, req CreatePostRequest, authorID primitive.ObjectID) (*PostResponse, error)
	GetPost(ctx context.Context, id primitive.ObjectID, viewerID *primitive.ObjectID) (*PostResponse, error)
	UpdatePost(ctx context.Context, id primitive.ObjectID, req UpdatePostRequest, userID primitive.ObjectID) (*PostResponse, error)
	DeletePost(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error

	// Query operations
	GetPosts(ctx context.Context, filter PostFilter, pagination PostPagination, viewerID *primitive.ObjectID) (*PostListResponse, error)
	GetUserPosts(ctx context.Context, userID primitive.ObjectID, pagination PostPagination, viewerID *primitive.ObjectID) (*PostListResponse, error)
	GetPostsByCategory(ctx context.Context, category string, pagination PostPagination, viewerID *primitive.ObjectID) (*PostListResponse, error)

	// Engagement operations
	LikePost(ctx context.Context, postID, userID primitive.ObjectID) error
	UnlikePost(ctx context.Context, postID, userID primitive.ObjectID) error

	// Search and discovery
	SearchPosts(ctx context.Context, query string, filter PostFilter, pagination PostPagination, viewerID *primitive.ObjectID) (*PostListResponse, error)
	GetPopularPosts(ctx context.Context, limit int, timeframe string, viewerID *primitive.ObjectID) (*PostListResponse, error)
	GetTrendingTags(ctx context.Context, limit int) ([]string, error)

	// Analytics
	GetPostAnalytics(ctx context.Context, postID primitive.ObjectID, userID primitive.ObjectID) (*PostAnalytics, error)
	GetUserPostStats(ctx context.Context, userID primitive.ObjectID) (*UserPostStats, error)

	// Moderation
	ReportPost(ctx context.Context, postID, reporterID primitive.ObjectID, reason string) error

	// Validation
	ValidatePostCategory(category string) error
	ValidateMediaLinks(mediaLinks []MediaLink) error
}

// PostListResponse represents a paginated list of posts
type PostListResponse struct {
	Posts      []PostResponse `json:"posts"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
	HasNext    bool           `json:"hasNext"`
	HasPrev    bool           `json:"hasPrev"`
}

// PostAnalytics represents detailed analytics for a post
type PostAnalytics struct {
	PostID           primitive.ObjectID `json:"postId"`
	ViewsCount       int                `json:"viewsCount"`
	LikesCount       int                `json:"likesCount"`
	CommentsCount    int                `json:"commentsCount"`
	SharesCount      int                `json:"sharesCount"`
	EngagementRate   float64            `json:"engagementRate"`
	ViewsByDay       []DayStats         `json:"viewsByDay"`
	LikesByDay       []DayStats         `json:"likesByDay"`
	TopReferrers     []string           `json:"topReferrers"`
	AudienceInsights AudienceInsights   `json:"audienceInsights"`
}

// UserPostStats represents statistics for a user's posts
type UserPostStats struct {
	UserID            primitive.ObjectID `json:"userId"`
	TotalPosts        int                `json:"totalPosts"`
	TotalViews        int                `json:"totalViews"`
	TotalLikes        int                `json:"totalLikes"`
	TotalComments     int                `json:"totalComments"`
	AverageEngagement float64            `json:"averageEngagement"`
	PopularCategories []CategoryStats    `json:"popularCategories"`
	PostsByMonth      []MonthStats       `json:"postsByMonth"`
}

// DayStats represents statistics for a specific day
type DayStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Views    int    `json:"views"`
	Likes    int    `json:"likes"`
}

// MonthStats represents statistics for a month
type MonthStats struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// AudienceInsights represents insights about post audience
type AudienceInsights struct {
	TopViewerRoles     []string           `json:"topViewerRoles"`
	EngagementByRole   map[string]float64 `json:"engagementByRole"`
	GeographicSpread   []string           `json:"geographicSpread"`
	PeakEngagementTime string             `json:"peakEngagementTime"`
}
