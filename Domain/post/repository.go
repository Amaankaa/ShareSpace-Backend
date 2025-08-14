package postpkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostRepository defines the interface for post data operations
type PostRepository interface {
	// Basic CRUD operations
	CreatePost(ctx context.Context, post Post) (*Post, error)
	GetPostByID(ctx context.Context, id primitive.ObjectID) (*Post, error)
	UpdatePost(ctx context.Context, id primitive.ObjectID, updates Post) (*Post, error)
	DeletePost(ctx context.Context, id primitive.ObjectID) error
	
	// Query operations
	GetPosts(ctx context.Context, filter PostFilter, pagination PostPagination) ([]Post, int64, error)
	GetPostsByAuthor(ctx context.Context, authorID primitive.ObjectID, pagination PostPagination) ([]Post, int64, error)
	GetPostsByCategory(ctx context.Context, category string, pagination PostPagination) ([]Post, int64, error)
	GetPostsByTag(ctx context.Context, tag string, pagination PostPagination) ([]Post, int64, error)
	
	// Engagement operations
	LikePost(ctx context.Context, postID, userID primitive.ObjectID) error
	UnlikePost(ctx context.Context, postID, userID primitive.ObjectID) error
	IsPostLikedByUser(ctx context.Context, postID, userID primitive.ObjectID) (bool, error)
	IncrementViewCount(ctx context.Context, postID primitive.ObjectID) error
	UpdateCommentsCount(ctx context.Context, postID primitive.ObjectID, increment int) error
	
	// Search operations
	SearchPosts(ctx context.Context, query string, filter PostFilter, pagination PostPagination) ([]Post, int64, error)
	
	// Analytics operations
	GetPopularPosts(ctx context.Context, limit int, timeframe string) ([]Post, error)
	GetTrendingTags(ctx context.Context, limit int) ([]string, error)
	GetPostStats(ctx context.Context, postID primitive.ObjectID) (*PostStats, error)
	
	// Moderation operations
	ReportPost(ctx context.Context, postID primitive.ObjectID) error
	HidePost(ctx context.Context, postID primitive.ObjectID) error
	UnhidePost(ctx context.Context, postID primitive.ObjectID) error
}

// PostStats represents analytics data for a post
type PostStats struct {
	PostID        primitive.ObjectID `json:"postId"`
	ViewsCount    int                `json:"viewsCount"`
	LikesCount    int                `json:"likesCount"`
	CommentsCount int                `json:"commentsCount"`
	SharesCount   int                `json:"sharesCount"`
	CreatedAt     string             `json:"createdAt"`
	Category      string             `json:"category"`
	Tags          []string           `json:"tags"`
}
