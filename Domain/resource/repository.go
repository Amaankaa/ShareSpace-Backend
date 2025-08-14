package resourcepkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResourceRepository defines the interface for resource data operations
type ResourceRepository interface {
	// Basic CRUD operations
	CreateResource(ctx context.Context, resource Resource) (*Resource, error)
	GetResourceByID(ctx context.Context, id primitive.ObjectID) (*Resource, error)
	UpdateResource(ctx context.Context, id primitive.ObjectID, updates Resource) (*Resource, error)
	DeleteResource(ctx context.Context, id primitive.ObjectID) error
	
	// Query operations
	GetResources(ctx context.Context, filter ResourceFilter, pagination ResourcePagination) ([]Resource, int64, error)
	GetResourcesByCreator(ctx context.Context, creatorID primitive.ObjectID, pagination ResourcePagination) ([]Resource, int64, error)
	GetResourcesByType(ctx context.Context, resourceType string, pagination ResourcePagination) ([]Resource, int64, error)
	GetResourcesByCategory(ctx context.Context, category string, pagination ResourcePagination) ([]Resource, int64, error)
	GetResourcesByTag(ctx context.Context, tag string, pagination ResourcePagination) ([]Resource, int64, error)
	
	// Engagement operations
	LikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	UnlikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	BookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	UnbookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	IsResourceLikedByUser(ctx context.Context, resourceID, userID primitive.ObjectID) (bool, error)
	IsResourceBookmarkedByUser(ctx context.Context, resourceID, userID primitive.ObjectID) (bool, error)
	IncrementViewCount(ctx context.Context, resourceID primitive.ObjectID) error
	IncrementShareCount(ctx context.Context, resourceID primitive.ObjectID) error
	
	// Rating operations
	RateResource(ctx context.Context, resourceID, userID primitive.ObjectID, rating float64) error
	GetResourceRating(ctx context.Context, resourceID primitive.ObjectID) (float64, int, error)
	GetUserRatingForResource(ctx context.Context, resourceID, userID primitive.ObjectID) (float64, error)
	
	// Search operations
	SearchResources(ctx context.Context, query string, filter ResourceFilter, pagination ResourcePagination) ([]Resource, int64, error)
	
	// Analytics operations
	GetPopularResources(ctx context.Context, limit int, timeframe string) ([]Resource, error)
	GetTrendingResources(ctx context.Context, limit int) ([]Resource, error)
	GetResourceStats(ctx context.Context, resourceID primitive.ObjectID) (*ResourceStats, error)
	GetTopRatedResources(ctx context.Context, limit int, category string) ([]Resource, error)
	
	// Verification operations
	VerifyResource(ctx context.Context, resourceID, verifierID primitive.ObjectID) error
	UnverifyResource(ctx context.Context, resourceID primitive.ObjectID) error
	GetUnverifiedResources(ctx context.Context, pagination ResourcePagination) ([]Resource, int64, error)
	
	// Moderation operations
	ReportResource(ctx context.Context, resourceID primitive.ObjectID) error
	HideResource(ctx context.Context, resourceID primitive.ObjectID) error
	UnhideResource(ctx context.Context, resourceID primitive.ObjectID) error
	
	// Deadline operations
	GetResourcesWithUpcomingDeadlines(ctx context.Context, days int, pagination ResourcePagination) ([]Resource, int64, error)
	GetExpiredResources(ctx context.Context, pagination ResourcePagination) ([]Resource, int64, error)
	
	// User-specific operations
	GetUserBookmarkedResources(ctx context.Context, userID primitive.ObjectID, pagination ResourcePagination) ([]Resource, int64, error)
	GetUserLikedResources(ctx context.Context, userID primitive.ObjectID, pagination ResourcePagination) ([]Resource, int64, error)
}

// ResourceStats represents analytics data for a resource
type ResourceStats struct {
	ResourceID     primitive.ObjectID `json:"resourceId"`
	ViewsCount     int                `json:"viewsCount"`
	LikesCount     int                `json:"likesCount"`
	BookmarksCount int                `json:"bookmarksCount"`
	SharesCount    int                `json:"sharesCount"`
	Rating         float64            `json:"rating"`
	RatingCount    int                `json:"ratingCount"`
	CreatedAt      string             `json:"createdAt"`
	Type           string             `json:"type"`
	Category       string             `json:"category"`
	Tags           []string           `json:"tags"`
	EngagementRate float64            `json:"engagementRate"`
}
