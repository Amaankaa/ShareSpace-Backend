package usecases

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostUsecase struct {
	postRepo postpkg.PostRepository
	userRepo userpkg.IUserRepository
}

func NewPostUsecase(
	postRepo postpkg.PostRepository,
	userRepo userpkg.IUserRepository,
) *PostUsecase {
	return &PostUsecase{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post with validation
func (uc *PostUsecase) CreatePost(ctx context.Context, req postpkg.CreatePostRequest, authorID primitive.ObjectID) (*postpkg.PostResponse, error) {
	// Validate category
	if err := uc.ValidatePostCategory(req.Category); err != nil {
		return nil, err
	}

	// Validate media links
	if err := uc.ValidateMediaLinks(req.MediaLinks); err != nil {
		return nil, err
	}

	// Get author information
	author, err := uc.userRepo.FindByID(ctx, authorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Create post entity
	post := postpkg.Post{
		AuthorID:    authorID,
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Category:    req.Category,
		Tags:        uc.normalizeTags(req.Tags),
		MediaLinks:  req.MediaLinks,
		IsAnonymous: req.IsAnonymous,
	}

	// Save post
	createdPost, err := uc.postRepo.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Convert to response
	return uc.convertToPostResponse(*createdPost, &author, false), nil
}

// GetPost retrieves a single post and increments view count
func (uc *PostUsecase) GetPost(ctx context.Context, id primitive.ObjectID, viewerID *primitive.ObjectID) (*postpkg.PostResponse, error) {
	// Get post
	post, err := uc.postRepo.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get author information
	author, err := uc.userRepo.FindByID(ctx, post.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Increment view count (don't block on errors)
	_ = uc.postRepo.IncrementViewCount(ctx, id)

	// Check if viewer has liked the post
	var isLikedByUser bool
	if viewerID != nil {
		isLikedByUser, _ = uc.postRepo.IsPostLikedByUser(ctx, id, *viewerID)
	}

	return uc.convertToPostResponse(*post, &author, isLikedByUser), nil
}

// UpdatePost updates an existing post (only by author)
func (uc *PostUsecase) UpdatePost(ctx context.Context, id primitive.ObjectID, req postpkg.UpdatePostRequest, userID primitive.ObjectID) (*postpkg.PostResponse, error) {
	// Get existing post
	existingPost, err := uc.postRepo.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user is the author
	if existingPost.AuthorID != userID {
		return nil, errors.New("unauthorized: only the author can update this post")
	}

	// Validate updates
	if req.Category != "" {
		if err := uc.ValidatePostCategory(req.Category); err != nil {
			return nil, err
		}
	}

	if req.MediaLinks != nil {
		if err := uc.ValidateMediaLinks(req.MediaLinks); err != nil {
			return nil, err
		}
	}

	// Build update object
	updates := postpkg.Post{}
	if req.Title != "" {
		updates.Title = strings.TrimSpace(req.Title)
	}
	if req.Content != "" {
		updates.Content = strings.TrimSpace(req.Content)
	}
	if req.Category != "" {
		updates.Category = req.Category
	}
	if req.Tags != nil {
		updates.Tags = uc.normalizeTags(req.Tags)
	}
	if req.MediaLinks != nil {
		updates.MediaLinks = req.MediaLinks
	}

	// Update post
	updatedPost, err := uc.postRepo.UpdatePost(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Get author information
	author, err := uc.userRepo.FindByID(ctx, updatedPost.AuthorID.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	return uc.convertToPostResponse(*updatedPost, &author, false), nil
}

// DeletePost deletes a post (only by author)
func (uc *PostUsecase) DeletePost(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Get existing post
	existingPost, err := uc.postRepo.GetPostByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user is the author
	if existingPost.AuthorID != userID {
		return errors.New("unauthorized: only the author can delete this post")
	}

	// Delete post
	return uc.postRepo.DeletePost(ctx, id)
}

// GetPosts retrieves posts with filtering and pagination
func (uc *PostUsecase) GetPosts(ctx context.Context, filter postpkg.PostFilter, pagination postpkg.PostPagination, viewerID *primitive.ObjectID) (*postpkg.PostListResponse, error) {
	// Set default pagination values
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 20
	}
	if pagination.SortBy == "" {
		pagination.SortBy = "createdAt"
	}
	if pagination.SortOrder == "" {
		pagination.SortOrder = "desc"
	}

	// Get posts
	posts, total, err := uc.postRepo.GetPosts(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	// Convert to responses
	postResponses, err := uc.convertToPostResponses(ctx, posts, viewerID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	hasNext := pagination.Page < totalPages
	hasPrev := pagination.Page > 1

	return &postpkg.PostListResponse{
		Posts:      postResponses,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetUserPosts retrieves posts by a specific user
func (uc *PostUsecase) GetUserPosts(ctx context.Context, userID primitive.ObjectID, pagination postpkg.PostPagination, viewerID *primitive.ObjectID) (*postpkg.PostListResponse, error) {
	filter := postpkg.PostFilter{AuthorID: userID.Hex()}
	return uc.GetPosts(ctx, filter, pagination, viewerID)
}

// GetPostsByCategory retrieves posts by category
func (uc *PostUsecase) GetPostsByCategory(ctx context.Context, category string, pagination postpkg.PostPagination, viewerID *primitive.ObjectID) (*postpkg.PostListResponse, error) {
	filter := postpkg.PostFilter{Category: category}
	return uc.GetPosts(ctx, filter, pagination, viewerID)
}

// LikePost adds a like to a post
func (uc *PostUsecase) LikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	// Check if post exists
	_, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check if already liked
	isLiked, err := uc.postRepo.IsPostLikedByUser(ctx, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to check like status: %w", err)
	}

	if isLiked {
		return errors.New("post already liked by user")
	}

	return uc.postRepo.LikePost(ctx, postID, userID)
}

// UnlikePost removes a like from a post
func (uc *PostUsecase) UnlikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	// Check if post exists
	_, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check if actually liked
	isLiked, err := uc.postRepo.IsPostLikedByUser(ctx, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to check like status: %w", err)
	}

	if !isLiked {
		return errors.New("post not liked by user")
	}

	return uc.postRepo.UnlikePost(ctx, postID, userID)
}

// SearchPosts searches posts by query
func (uc *PostUsecase) SearchPosts(ctx context.Context, query string, filter postpkg.PostFilter, pagination postpkg.PostPagination, viewerID *primitive.ObjectID) (*postpkg.PostListResponse, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// Set default pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 20
	}

	// Search posts
	posts, total, err := uc.postRepo.SearchPosts(ctx, query, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	// Convert to responses
	postResponses, err := uc.convertToPostResponses(ctx, posts, viewerID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	hasNext := pagination.Page < totalPages
	hasPrev := pagination.Page > 1

	return &postpkg.PostListResponse{
		Posts:      postResponses,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

// GetPopularPosts retrieves popular posts
func (uc *PostUsecase) GetPopularPosts(ctx context.Context, limit int, timeframe string, viewerID *primitive.ObjectID) (*postpkg.PostListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	posts, err := uc.postRepo.GetPopularPosts(ctx, limit, timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular posts: %w", err)
	}

	postResponses, err := uc.convertToPostResponses(ctx, posts, viewerID)
	if err != nil {
		return nil, err
	}

	return &postpkg.PostListResponse{
		Posts:      postResponses,
		Total:      int64(len(postResponses)),
		Page:       1,
		PageSize:   limit,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}, nil
}

// GetTrendingTags retrieves trending tags
func (uc *PostUsecase) GetTrendingTags(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return uc.postRepo.GetTrendingTags(ctx, limit)
}

// ReportPost reports a post for moderation
func (uc *PostUsecase) ReportPost(ctx context.Context, postID, reporterID primitive.ObjectID, reason string) error {
	// Check if post exists
	_, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return err
	}

	// TODO: Store report details with reason and reporter
	return uc.postRepo.ReportPost(ctx, postID)
}

// ValidatePostCategory validates if a category is allowed
func (uc *PostUsecase) ValidatePostCategory(category string) error {
	if !slices.Contains(postpkg.PostCategories, category) {
		return fmt.Errorf("invalid category: %s", category)
	}
	return nil
}

// ValidateMediaLinks validates media links
func (uc *PostUsecase) ValidateMediaLinks(mediaLinks []postpkg.MediaLink) error {
	allowedTypes := []string{postpkg.MediaTypeImage, postpkg.MediaTypeVideo, postpkg.MediaTypeDocument, postpkg.MediaTypeLink}

	for _, link := range mediaLinks {
		if !slices.Contains(allowedTypes, link.Type) {
			return fmt.Errorf("invalid media type: %s", link.Type)
		}

		if strings.TrimSpace(link.URL) == "" {
			return errors.New("media URL cannot be empty")
		}
	}

	return nil
}

// Helper functions

// normalizeTags normalizes and deduplicates tags
func (uc *PostUsecase) normalizeTags(tags []string) []string {
	var normalized []string
	seen := make(map[string]bool)

	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" && !seen[tag] {
			normalized = append(normalized, tag)
			seen[tag] = true
		}
	}

	return normalized
}

// convertToPostResponse converts a post entity to response format
func (uc *PostUsecase) convertToPostResponse(post postpkg.Post, author *userpkg.User, isLikedByUser bool) *postpkg.PostResponse {
	// Handle anonymous posts
	authorInfo := postpkg.AuthorInfo{
		ID:             author.ID,
		DisplayName:    author.DisplayName,
		ProfilePicture: author.ProfilePicture,
		IsMentor:       author.IsMentor,
		IsAnonymous:    post.IsAnonymous,
	}

	if post.IsAnonymous {
		authorInfo.DisplayName = "Anonymous"
		authorInfo.ProfilePicture = ""
	}

	return &postpkg.PostResponse{
		ID:            post.ID,
		Author:        authorInfo,
		Title:         post.Title,
		Content:       post.Content,
		Category:      post.Category,
		Tags:          post.Tags,
		MediaLinks:    post.MediaLinks,
		IsAnonymous:   post.IsAnonymous,
		LikesCount:    post.LikesCount,
		CommentsCount: post.CommentsCount,
		ViewsCount:    post.ViewsCount,
		IsLikedByUser: isLikedByUser,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
	}
}

// convertToPostResponses converts multiple posts to response format
func (uc *PostUsecase) convertToPostResponses(ctx context.Context, posts []postpkg.Post, viewerID *primitive.ObjectID) ([]postpkg.PostResponse, error) {
	var responses []postpkg.PostResponse

	for _, post := range posts {
		// Get author
		author, err := uc.userRepo.FindByID(ctx, post.AuthorID.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get author for post %s: %w", post.ID.Hex(), err)
		}

		// Check if viewer has liked the post
		var isLikedByUser bool
		if viewerID != nil {
			isLikedByUser, _ = uc.postRepo.IsPostLikedByUser(ctx, post.ID, *viewerID)
		}

		response := uc.convertToPostResponse(post, &author, isLikedByUser)
		responses = append(responses, *response)
	}

	return responses, nil
}

// GetPostAnalytics retrieves analytics for a post (only for author)
func (uc *PostUsecase) GetPostAnalytics(ctx context.Context, postID primitive.ObjectID, userID primitive.ObjectID) (*postpkg.PostAnalytics, error) {
	// Get post to check ownership
	post, err := uc.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check if user is the author
	if post.AuthorID != userID {
		return nil, errors.New("unauthorized: only the author can view analytics")
	}

	// Get basic stats
	stats, err := uc.postRepo.GetPostStats(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post stats: %w", err)
	}

	// Calculate engagement rate
	totalEngagement := stats.LikesCount + stats.CommentsCount
	engagementRate := 0.0
	if stats.ViewsCount > 0 {
		engagementRate = float64(totalEngagement) / float64(stats.ViewsCount) * 100
	}

	return &postpkg.PostAnalytics{
		PostID:         stats.PostID,
		ViewsCount:     stats.ViewsCount,
		LikesCount:     stats.LikesCount,
		CommentsCount:  stats.CommentsCount,
		SharesCount:    stats.SharesCount,
		EngagementRate: engagementRate,
		// TODO: Implement detailed analytics (views by day, etc.)
		ViewsByDay:       []postpkg.DayStats{},
		LikesByDay:       []postpkg.DayStats{},
		TopReferrers:     []string{},
		AudienceInsights: postpkg.AudienceInsights{},
	}, nil
}

// GetUserPostStats retrieves statistics for a user's posts
func (uc *PostUsecase) GetUserPostStats(ctx context.Context, userID primitive.ObjectID) (*postpkg.UserPostStats, error) {
	// Get user's posts
	pagination := postpkg.PostPagination{Page: 1, PageSize: 1000} // Get all posts for stats
	posts, total, err := uc.postRepo.GetPostsByAuthor(ctx, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get user posts: %w", err)
	}

	// Calculate statistics
	totalViews := 0
	totalLikes := 0
	totalComments := 0
	categoryStats := make(map[string]*postpkg.CategoryStats)

	for _, post := range posts {
		totalViews += post.ViewsCount
		totalLikes += post.LikesCount
		totalComments += post.CommentsCount

		// Category stats
		if stat, exists := categoryStats[post.Category]; exists {
			stat.Count++
			stat.Views += post.ViewsCount
			stat.Likes += post.LikesCount
		} else {
			categoryStats[post.Category] = &postpkg.CategoryStats{
				Category: post.Category,
				Count:    1,
				Views:    post.ViewsCount,
				Likes:    post.LikesCount,
			}
		}
	}

	// Convert category stats to slice
	var popularCategories []postpkg.CategoryStats
	for _, stat := range categoryStats {
		popularCategories = append(popularCategories, *stat)
	}

	// Calculate average engagement
	averageEngagement := 0.0
	if totalViews > 0 {
		averageEngagement = float64(totalLikes+totalComments) / float64(totalViews) * 100
	}

	return &postpkg.UserPostStats{
		UserID:            userID,
		TotalPosts:        int(total),
		TotalViews:        totalViews,
		TotalLikes:        totalLikes,
		TotalComments:     totalComments,
		AverageEngagement: averageEngagement,
		PopularCategories: popularCategories,
		PostsByMonth:      []postpkg.MonthStats{}, // TODO: Implement monthly breakdown
	}, nil
}
