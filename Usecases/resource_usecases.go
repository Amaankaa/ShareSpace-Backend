package usecases

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	resourcepkg "github.com/Amaankaa/Blog-Starter-Project/Domain/resource"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceUsecase struct {
	resourceRepo resourcepkg.ResourceRepository
	userRepo     userpkg.IUserRepository
}

func NewResourceUsecase(resourceRepo resourcepkg.ResourceRepository, userRepo userpkg.IUserRepository) *ResourceUsecase {
	return &ResourceUsecase{resourceRepo: resourceRepo, userRepo: userRepo}
}

// Core
func (uc *ResourceUsecase) CreateResource(ctx context.Context, req resourcepkg.CreateResourceRequest, creatorID primitive.ObjectID) (*resourcepkg.ResourceResponse, error) {
	if err := uc.ValidateResourceType(req.Type); err != nil {
		return nil, err
	}
	if err := uc.ValidateResourceCategory(req.Category); err != nil {
		return nil, err
	}
	if err := uc.ValidateDifficulty(req.Difficulty); err != nil {
		return nil, err
	}
	if err := uc.ValidateAttachments(req.Attachments); err != nil {
		return nil, err
	}

	creator, err := uc.userRepo.FindByID(ctx, creatorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	res := resourcepkg.Resource{
		CreatorID:      creatorID,
		Title:          strings.TrimSpace(req.Title),
		Description:    strings.TrimSpace(req.Description),
		Content:        strings.TrimSpace(req.Content),
		Type:           req.Type,
		Category:       req.Category,
		Tags:           normalizeTags(req.Tags),
		ExternalURL:    req.ExternalURL,
		Attachments:    req.Attachments,
		Difficulty:     req.Difficulty,
		EstimatedTime:  req.EstimatedTime,
		Prerequisites:  req.Prerequisites,
		Deadline:       req.Deadline,
		ApplicationURL: req.ApplicationURL,
		Eligibility:    req.Eligibility,
		Amount:         req.Amount,
	}

	created, err := uc.resourceRepo.CreateResource(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	return uc.convertToResourceResponse(ctx, *created, &creator, nil)
}

func (uc *ResourceUsecase) GetResource(ctx context.Context, id primitive.ObjectID, viewerID *primitive.ObjectID) (*resourcepkg.ResourceResponse, error) {
	res, err := uc.resourceRepo.GetResourceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	creator, err := uc.userRepo.FindByID(ctx, res.CreatorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}
	_ = uc.resourceRepo.IncrementViewCount(ctx, id)
	return uc.convertToResourceResponse(ctx, *res, &creator, viewerID)
}

func (uc *ResourceUsecase) UpdateResource(ctx context.Context, id primitive.ObjectID, req resourcepkg.UpdateResourceRequest, userID primitive.ObjectID) (*resourcepkg.ResourceResponse, error) {
	existing, err := uc.resourceRepo.GetResourceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.CreatorID != userID {
		return nil, errors.New("unauthorized: only the creator can update this resource")
	}
	if req.Type != "" {
		if err := uc.ValidateResourceType(req.Type); err != nil {
			return nil, err
		}
	}
	if req.Category != "" {
		if err := uc.ValidateResourceCategory(req.Category); err != nil {
			return nil, err
		}
	}
	if req.Difficulty != "" {
		if err := uc.ValidateDifficulty(req.Difficulty); err != nil {
			return nil, err
		}
	}
	if req.Attachments != nil {
		if err := uc.ValidateAttachments(req.Attachments); err != nil {
			return nil, err
		}
	}

	updates := resourcepkg.Resource{}
	if req.Title != "" {
		updates.Title = strings.TrimSpace(req.Title)
	}
	if req.Description != "" {
		updates.Description = strings.TrimSpace(req.Description)
	}
	if req.Content != "" {
		updates.Content = strings.TrimSpace(req.Content)
	}
	if req.Type != "" {
		updates.Type = req.Type
	}
	if req.Category != "" {
		updates.Category = req.Category
	}
	if req.Tags != nil {
		updates.Tags = normalizeTags(req.Tags)
	}
	if req.ExternalURL != "" {
		updates.ExternalURL = req.ExternalURL
	}
	if req.Attachments != nil {
		updates.Attachments = req.Attachments
	}
	if req.Difficulty != "" {
		updates.Difficulty = req.Difficulty
	}
	if req.EstimatedTime != "" {
		updates.EstimatedTime = req.EstimatedTime
	}
	if req.Prerequisites != nil {
		updates.Prerequisites = req.Prerequisites
	}
	if req.Deadline != nil {
		updates.Deadline = req.Deadline
	}
	if req.ApplicationURL != "" {
		updates.ApplicationURL = req.ApplicationURL
	}
	if req.Eligibility != nil {
		updates.Eligibility = req.Eligibility
	}
	if req.Amount != "" {
		updates.Amount = req.Amount
	}

	updated, err := uc.resourceRepo.UpdateResource(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}
	creator, err := uc.userRepo.FindByID(ctx, updated.CreatorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}
	return uc.convertToResourceResponse(ctx, *updated, &creator, nil)
}

func (uc *ResourceUsecase) DeleteResource(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	existing, err := uc.resourceRepo.GetResourceByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.CreatorID != userID {
		return errors.New("unauthorized: only the creator can delete this resource")
	}
	return uc.resourceRepo.DeleteResource(ctx, id)
}

// Lists
func (uc *ResourceUsecase) GetResources(ctx context.Context, filter resourcepkg.ResourceFilter, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetResources(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetUserResources(ctx context.Context, userID primitive.ObjectID, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetResourcesByCreator(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetResourcesByType(ctx context.Context, resourceType string, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetResourcesByType(ctx, resourceType, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources by type: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetResourcesByCategory(ctx context.Context, category string, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetResourcesByCategory(ctx, category, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources by category: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

// Engagement
func (uc *ResourceUsecase) LikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error {
	// Ensure resource exists and not already liked
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	liked, err := uc.resourceRepo.IsResourceLikedByUser(ctx, resourceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check like status: %w", err)
	}
	if liked {
		return errors.New("resource already liked by user")
	}
	return uc.resourceRepo.LikeResource(ctx, resourceID, userID)
}

func (uc *ResourceUsecase) UnlikeResource(ctx context.Context, resourceID, userID primitive.ObjectID) error {
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	liked, err := uc.resourceRepo.IsResourceLikedByUser(ctx, resourceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check like status: %w", err)
	}
	if !liked {
		return errors.New("resource not liked by user")
	}
	return uc.resourceRepo.UnlikeResource(ctx, resourceID, userID)
}

func (uc *ResourceUsecase) BookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error {
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	booked, err := uc.resourceRepo.IsResourceBookmarkedByUser(ctx, resourceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check bookmark status: %w", err)
	}
	if booked {
		return errors.New("resource already bookmarked by user")
	}
	return uc.resourceRepo.BookmarkResource(ctx, resourceID, userID)
}

func (uc *ResourceUsecase) UnbookmarkResource(ctx context.Context, resourceID, userID primitive.ObjectID) error {
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	booked, err := uc.resourceRepo.IsResourceBookmarkedByUser(ctx, resourceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check bookmark status: %w", err)
	}
	if !booked {
		return errors.New("resource not bookmarked by user")
	}
	return uc.resourceRepo.UnbookmarkResource(ctx, resourceID, userID)
}

func (uc *ResourceUsecase) RateResource(ctx context.Context, resourceID, userID primitive.ObjectID, rating float64) error {
	if rating < 0 || rating > 5 {
		return errors.New("rating must be between 0 and 5")
	}
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	return uc.resourceRepo.RateResource(ctx, resourceID, userID, rating)
}

// Search & discovery
func (uc *ResourceUsecase) SearchResources(ctx context.Context, query string, filter resourcepkg.ResourceFilter, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query cannot be empty")
	}
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.SearchResources(ctx, query, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to search resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetPopularResources(ctx context.Context, limit int, timeframe string, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := uc.resourceRepo.GetPopularResources(ctx, limit, timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return &resourcepkg.ResourceListResponse{Resources: resp, Total: int64(len(resp)), Page: 1, PageSize: limit, TotalPages: 1}, nil
}

func (uc *ResourceUsecase) GetTrendingResources(ctx context.Context, limit int, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := uc.resourceRepo.GetTrendingResources(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return &resourcepkg.ResourceListResponse{Resources: resp, Total: int64(len(resp)), Page: 1, PageSize: limit, TotalPages: 1}, nil
}

func (uc *ResourceUsecase) GetTopRatedResources(ctx context.Context, limit int, category string, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := uc.resourceRepo.GetTopRatedResources(ctx, limit, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get top rated resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return &resourcepkg.ResourceListResponse{Resources: resp, Total: int64(len(resp)), Page: 1, PageSize: limit, TotalPages: 1}, nil
}

func (uc *ResourceUsecase) GetRecommendedResources(ctx context.Context, userID primitive.ObjectID, limit int) (*resourcepkg.ResourceListResponse, error) {
	// Placeholder recommendation: popular resources
	return uc.GetPopularResources(ctx, limit, "week", nil)
}

// User-specific
func (uc *ResourceUsecase) GetUserBookmarkedResources(ctx context.Context, userID primitive.ObjectID, pagination resourcepkg.ResourcePagination) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetUserBookmarkedResources(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarked resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, &userID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetUserLikedResources(ctx context.Context, userID primitive.ObjectID, pagination resourcepkg.ResourcePagination) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetUserLikedResources(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked resources: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, &userID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

// Deadlines
func (uc *ResourceUsecase) GetUpcomingOpportunities(ctx context.Context, days int, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	items, total, err := uc.resourceRepo.GetResourcesWithUpcomingDeadlines(ctx, days, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming opportunities: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

func (uc *ResourceUsecase) GetResourcesWithDeadlines(ctx context.Context, pagination resourcepkg.ResourcePagination, viewerID *primitive.ObjectID) (*resourcepkg.ResourceListResponse, error) {
	setDefaults(&pagination)
	filter := resourcepkg.ResourceFilter{HasDeadline: ptrBool(true)}
	items, total, err := uc.resourceRepo.GetResources(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources with deadlines: %w", err)
	}
	resp, err := uc.convertMany(ctx, items, viewerID)
	if err != nil {
		return nil, err
	}
	return listResponse(resp, total, pagination), nil
}

// Analytics
func (uc *ResourceUsecase) GetResourceAnalytics(ctx context.Context, resourceID primitive.ObjectID, userID primitive.ObjectID) (*resourcepkg.ResourceAnalytics, error) {
	res, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	if res.CreatorID != userID {
		return nil, errors.New("unauthorized: only the creator can view analytics")
	}
	stats, err := uc.resourceRepo.GetResourceStats(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource stats: %w", err)
	}
	// Basic mapping
	return &resourcepkg.ResourceAnalytics{
		ResourceID:       stats.ResourceID,
		ViewsCount:       stats.ViewsCount,
		LikesCount:       stats.LikesCount,
		BookmarksCount:   stats.BookmarksCount,
		SharesCount:      stats.SharesCount,
		Rating:           stats.Rating,
		RatingCount:      stats.RatingCount,
		EngagementRate:   0,
		ViewsByDay:       []resourcepkg.DayStats{},
		LikesByDay:       []resourcepkg.DayStats{},
		BookmarksByDay:   []resourcepkg.DayStats{},
		TopReferrers:     []string{},
		AudienceInsights: resourcepkg.AudienceInsights{},
		QualityMetrics:   resourcepkg.QualityMetrics{},
	}, nil
}

func (uc *ResourceUsecase) GetUserResourceStats(ctx context.Context, userID primitive.ObjectID) (*resourcepkg.UserResourceStats, error) {
	// Aggregate via listing for now
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 1000}
	items, total, err := uc.resourceRepo.GetResourcesByCreator(ctx, userID, pg)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resources: %w", err)
	}
	var views, likes, bookmarks int
	for _, it := range items {
		views += it.ViewsCount
		likes += it.LikesCount
		bookmarks += it.BookmarksCount
	}
	avgEng := 0.0
	if views > 0 {
		avgEng = float64(likes+bookmarks) / float64(views) * 100
	}
	return &resourcepkg.UserResourceStats{
		UserID:                userID,
		TotalResources:        int(total),
		TotalViews:            views,
		TotalLikes:            likes,
		TotalBookmarks:        bookmarks,
		AverageEngagement:     avgEng,
		VerifiedResources:     0,
		PopularTypes:          []resourcepkg.TypeStats{},
		PopularCategories:     []resourcepkg.CategoryStats{},
		ResourcesByMonth:      []resourcepkg.MonthStats{},
		TopPerformingResource: nil,
	}, nil
}

// Moderation / Verification
func (uc *ResourceUsecase) ReportResource(ctx context.Context, resourceID, reporterID primitive.ObjectID, reason string) error {
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	// TODO: persist reporterID and reason somewhere
	return uc.resourceRepo.ReportResource(ctx, resourceID)
}

func (uc *ResourceUsecase) VerifyResource(ctx context.Context, resourceID, verifierID primitive.ObjectID) error {
	_, err := uc.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return err
	}
	return uc.resourceRepo.VerifyResource(ctx, resourceID, verifierID)
}

// Validation helpers
func (uc *ResourceUsecase) ValidateResourceType(resourceType string) error {
	if resourceType == "" {
		return nil
	}
	for _, t := range resourcepkg.ResourceTypes {
		if t == resourceType {
			return nil
		}
	}
	return fmt.Errorf("invalid resource type: %s", resourceType)
}

func (uc *ResourceUsecase) ValidateResourceCategory(category string) error {
	if category == "" {
		return nil
	}
	for _, c := range resourcepkg.ResourceCategories {
		if c == category {
			return nil
		}
	}
	return fmt.Errorf("invalid resource category: %s", category)
}

func (uc *ResourceUsecase) ValidateDifficulty(difficulty string) error {
	if difficulty == "" {
		return nil
	}
	for _, d := range resourcepkg.DifficultyLevels {
		if d == difficulty {
			return nil
		}
	}
	return fmt.Errorf("invalid difficulty: %s", difficulty)
}

func (uc *ResourceUsecase) ValidateAttachments(attachments []resourcepkg.Attachment) error {
	for _, a := range attachments {
		if strings.TrimSpace(a.URL) == "" {
			return errors.New("attachment URL cannot be empty")
		}
		if a.Type == "" {
			return errors.New("attachment type is required")
		}
	}
	return nil
}

// internal helpers
func (uc *ResourceUsecase) convertToResourceResponse(ctx context.Context, res resourcepkg.Resource, creator *userpkg.User, viewerID *primitive.ObjectID) (*resourcepkg.ResourceResponse, error) {
	var isLiked, isBookmarked bool
	if viewerID != nil {
		isLiked, _ = uc.resourceRepo.IsResourceLikedByUser(ctx, res.ID, *viewerID)
		isBookmarked, _ = uc.resourceRepo.IsResourceBookmarkedByUser(ctx, res.ID, *viewerID)
	}
	return &resourcepkg.ResourceResponse{
		ID: res.ID,
		Creator: resourcepkg.CreatorInfo{
			ID:             creator.ID,
			DisplayName:    creator.DisplayName,
			ProfilePicture: creator.ProfilePicture,
			IsMentor:       creator.IsMentor,
			IsVerified:     creator.IsVerified,
		},
		Title:              res.Title,
		Description:        res.Description,
		Content:            res.Content,
		Type:               res.Type,
		Category:           res.Category,
		Tags:               res.Tags,
		ExternalURL:        res.ExternalURL,
		Attachments:        res.Attachments,
		Difficulty:         res.Difficulty,
		EstimatedTime:      res.EstimatedTime,
		Prerequisites:      res.Prerequisites,
		Deadline:           res.Deadline,
		ApplicationURL:     res.ApplicationURL,
		Eligibility:        res.Eligibility,
		Amount:             res.Amount,
		ViewsCount:         res.ViewsCount,
		LikesCount:         res.LikesCount,
		BookmarksCount:     res.BookmarksCount,
		SharesCount:        res.SharesCount,
		IsLikedByUser:      isLiked,
		IsBookmarkedByUser: isBookmarked,
		IsVerified:         res.IsVerified,
		QualityScore:       res.QualityScore,
		Rating:             res.Rating,
		RatingCount:        res.RatingCount,
		CreatedAt:          res.CreatedAt,
		UpdatedAt:          res.UpdatedAt,
	}, nil
}

func (uc *ResourceUsecase) convertMany(ctx context.Context, items []resourcepkg.Resource, viewerID *primitive.ObjectID) ([]resourcepkg.ResourceResponse, error) {
	var out []resourcepkg.ResourceResponse
	for _, it := range items {
		creator, err := uc.userRepo.FindByID(ctx, it.CreatorID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get creator for resource %s: %w", it.ID.Hex(), err)
		}
		r, err := uc.convertToResourceResponse(ctx, it, &creator, viewerID)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

func listResponse(items []resourcepkg.ResourceResponse, total int64, pg resourcepkg.ResourcePagination) *resourcepkg.ResourceListResponse {
	totalPages := int(math.Ceil(float64(total) / float64(pg.PageSize)))
	return &resourcepkg.ResourceListResponse{
		Resources:  items,
		Total:      total,
		Page:       pg.Page,
		PageSize:   pg.PageSize,
		TotalPages: totalPages,
		HasNext:    pg.Page < totalPages,
		HasPrev:    pg.Page > 1,
	}
}

func setDefaults(pg *resourcepkg.ResourcePagination) {
	if pg.Page < 1 {
		pg.Page = 1
	}
	if pg.PageSize < 1 || pg.PageSize > 100 {
		pg.PageSize = 20
	}
	if pg.SortBy == "" {
		pg.SortBy = "createdAt"
	}
	if pg.SortOrder == "" {
		pg.SortOrder = "desc"
	}
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func ptrBool(v bool) *bool { return &v }
