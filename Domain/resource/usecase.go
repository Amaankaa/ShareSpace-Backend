package resourcepkg

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResourceUsecase defines the business logic interface for resources
type ResourceUsecase interface {
	// Core resource operations
	CreateResource(ctx context.Context, req CreateResourceRequest, creatorID primitive.ObjectID) (*ResourceResponse, error)
	GetResource(ctx context.Context, id primitive.ObjectID, viewerID *primitive.ObjectID) (*ResourceResponse, error)
	UpdateResource(ctx context.Context, id primitive.ObjectID, req UpdateResourceRequest, userID primitive.ObjectID) (*ResourceResponse, error)
	DeleteResource(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	
	// Query operations
	GetResources(ctx context.Context, filter ResourceFilter, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetUserResources(ctx context.Context, userID primitive.ObjectID, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetResourcesByType(ctx context.Context, resourceType string, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetResourcesByCategory(ctx context.Context, category string, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	
	// Engagement operations
	LikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	UnlikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	BookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	UnbookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error
	RateResource(ctx context.Context, resourceID, userID primitive.ObjectID, rating float64) error
	
	// Search and discovery
	SearchResources(ctx context.Context, query string, filter ResourceFilter, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetPopularResources(ctx context.Context, limit int, timeframe string, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetTrendingResources(ctx context.Context, limit int, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetTopRatedResources(ctx context.Context, limit int, category string, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetRecommendedResources(ctx context.Context, userID primitive.ObjectID, limit int) (*ResourceListResponse, error)
	
	// User-specific operations
	GetUserBookmarkedResources(ctx context.Context, userID primitive.ObjectID, pagination ResourcePagination) (*ResourceListResponse, error)
	GetUserLikedResources(ctx context.Context, userID primitive.ObjectID, pagination ResourcePagination) (*ResourceListResponse, error)
	
	// Deadline and opportunity management
	GetUpcomingOpportunities(ctx context.Context, days int, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	GetResourcesWithDeadlines(ctx context.Context, pagination ResourcePagination, viewerID *primitive.ObjectID) (*ResourceListResponse, error)
	
	// Analytics
	GetResourceAnalytics(ctx context.Context, resourceID primitive.ObjectID, userID primitive.ObjectID) (*ResourceAnalytics, error)
	GetUserResourceStats(ctx context.Context, userID primitive.ObjectID) (*UserResourceStats, error)
	
	// Moderation and verification
	ReportResource(ctx context.Context, resourceID, reporterID primitive.ObjectID, reason string) error
	VerifyResource(ctx context.Context, resourceID, verifierID primitive.ObjectID) error
	
	// Validation
	ValidateResourceType(resourceType string) error
	ValidateResourceCategory(category string) error
	ValidateDifficulty(difficulty string) error
	ValidateAttachments(attachments []Attachment) error
}

// ResourceListResponse represents a paginated list of resources
type ResourceListResponse struct {
	Resources  []ResourceResponse `json:"resources"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
	HasNext    bool               `json:"hasNext"`
	HasPrev    bool               `json:"hasPrev"`
}

// ResourceAnalytics represents detailed analytics for a resource
type ResourceAnalytics struct {
	ResourceID       primitive.ObjectID `json:"resourceId"`
	ViewsCount       int                `json:"viewsCount"`
	LikesCount       int                `json:"likesCount"`
	BookmarksCount   int                `json:"bookmarksCount"`
	SharesCount      int                `json:"sharesCount"`
	Rating           float64            `json:"rating"`
	RatingCount      int                `json:"ratingCount"`
	EngagementRate   float64            `json:"engagementRate"`
	ViewsByDay       []DayStats         `json:"viewsByDay"`
	LikesByDay       []DayStats         `json:"likesByDay"`
	BookmarksByDay   []DayStats         `json:"bookmarksByDay"`
	TopReferrers     []string           `json:"topReferrers"`
	AudienceInsights AudienceInsights   `json:"audienceInsights"`
	QualityMetrics   QualityMetrics     `json:"qualityMetrics"`
}

// UserResourceStats represents statistics for a user's resources
type UserResourceStats struct {
	UserID              primitive.ObjectID `json:"userId"`
	TotalResources      int                `json:"totalResources"`
	TotalViews          int                `json:"totalViews"`
	TotalLikes          int                `json:"totalLikes"`
	TotalBookmarks      int                `json:"totalBookmarks"`
	AverageRating       float64            `json:"averageRating"`
	AverageEngagement   float64            `json:"averageEngagement"`
	VerifiedResources   int                `json:"verifiedResources"`
	PopularTypes        []TypeStats        `json:"popularTypes"`
	PopularCategories   []CategoryStats    `json:"popularCategories"`
	ResourcesByMonth    []MonthStats       `json:"resourcesByMonth"`
	TopPerformingResource *ResourceResponse `json:"topPerformingResource"`
}

// DayStats represents statistics for a specific day
type DayStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// TypeStats represents statistics for a resource type
type TypeStats struct {
	Type      string  `json:"type"`
	Count     int     `json:"count"`
	Views     int     `json:"views"`
	Likes     int     `json:"likes"`
	Bookmarks int     `json:"bookmarks"`
	Rating    float64 `json:"rating"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category  string  `json:"category"`
	Count     int     `json:"count"`
	Views     int     `json:"views"`
	Likes     int     `json:"likes"`
	Bookmarks int     `json:"bookmarks"`
	Rating    float64 `json:"rating"`
}

// MonthStats represents statistics for a month
type MonthStats struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// AudienceInsights represents insights about resource audience
type AudienceInsights struct {
	TopViewerRoles     []string           `json:"topViewerRoles"`
	EngagementByRole   map[string]float64 `json:"engagementByRole"`
	GeographicSpread   []string           `json:"geographicSpread"`
	PeakEngagementTime string             `json:"peakEngagementTime"`
	UserRetention      float64            `json:"userRetention"`
}

// QualityMetrics represents quality assessment metrics
type QualityMetrics struct {
	ContentQuality    float64 `json:"contentQuality"`
	Usefulness        float64 `json:"usefulness"`
	Accuracy          float64 `json:"accuracy"`
	Completeness      float64 `json:"completeness"`
	OverallScore      float64 `json:"overallScore"`
	FeedbackCount     int     `json:"feedbackCount"`
	RecommendationRate float64 `json:"recommendationRate"`
}
