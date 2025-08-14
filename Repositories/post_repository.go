package repositories

import (
	"context"
	"fmt"
	"time"

	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostRepository struct {
	collection *mongo.Collection
}

func NewPostRepository(collection *mongo.Collection) *PostRepository {
	return &PostRepository{
		collection: collection,
	}
}

// CreatePost creates a new post in the database
func (r *PostRepository) CreatePost(ctx context.Context, post postpkg.Post) (*postpkg.Post, error) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.Status = postpkg.PostStatusActive
	post.LikesCount = 0
	post.CommentsCount = 0
	post.ViewsCount = 0
	post.LikedBy = []primitive.ObjectID{}

	_, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return &post, nil
}

// GetPostByID retrieves a post by its ID
func (r *PostRepository) GetPostByID(ctx context.Context, id primitive.ObjectID) (*postpkg.Post, error) {
	var post postpkg.Post
	filter := bson.M{"_id": id, "status": postpkg.PostStatusActive}

	err := r.collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

// UpdatePost updates an existing post
func (r *PostRepository) UpdatePost(ctx context.Context, id primitive.ObjectID, updates postpkg.Post) (*postpkg.Post, error) {
	updates.UpdatedAt = time.Now()

	updateDoc := bson.M{"$set": bson.M{}}
	if updates.Title != "" {
		updateDoc["$set"].(bson.M)["title"] = updates.Title
	}
	if updates.Content != "" {
		updateDoc["$set"].(bson.M)["content"] = updates.Content
	}
	if updates.Category != "" {
		updateDoc["$set"].(bson.M)["category"] = updates.Category
	}
	if updates.Tags != nil {
		updateDoc["$set"].(bson.M)["tags"] = updates.Tags
	}
	if updates.MediaLinks != nil {
		updateDoc["$set"].(bson.M)["mediaLinks"] = updates.MediaLinks
	}
	updateDoc["$set"].(bson.M)["updatedAt"] = updates.UpdatedAt

	filter := bson.M{"_id": id, "status": postpkg.PostStatusActive}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedPost postpkg.Post
	err := r.collection.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&updatedPost)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return &updatedPost, nil
}

// DeletePost soft deletes a post by updating its status
func (r *PostRepository) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$set": bson.M{
			"status":    postpkg.PostStatusDeleted,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// GetPosts retrieves posts with filtering and pagination
func (r *PostRepository) GetPosts(ctx context.Context, filter postpkg.PostFilter, pagination postpkg.PostPagination) ([]postpkg.Post, int64, error) {
	// Build filter
	mongoFilter := bson.M{"status": postpkg.PostStatusActive}

	if filter.Category != "" {
		mongoFilter["category"] = filter.Category
	}
	if filter.AuthorID != "" {
		authorID, err := primitive.ObjectIDFromHex(filter.AuthorID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid author ID: %w", err)
		}
		mongoFilter["authorId"] = authorID
	}
	if filter.Tag != "" {
		mongoFilter["tags"] = bson.M{"$in": []string{filter.Tag}}
	}
	if filter.Year != 0 {
		startOfYear := time.Date(filter.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(filter.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		mongoFilter["createdAt"] = bson.M{
			"$gte": startOfYear,
			"$lt":  endOfYear,
		}
	}
	if filter.IsAnonymous != nil {
		mongoFilter["isAnonymous"] = *filter.IsAnonymous
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Build sort options
	sortField := "createdAt"
	sortOrder := -1 // desc by default

	if pagination.SortBy != "" {
		sortField = pagination.SortBy
	}
	if pagination.SortOrder == "asc" {
		sortOrder = 1
	}

	// Build find options
	skip := int64((pagination.Page - 1) * pagination.PageSize)
	limit := int64(pagination.PageSize)

	findOptions := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(limit)

	// Execute query
	cursor, err := r.collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find posts: %w", err)
	}
	defer cursor.Close(ctx)

	var posts []postpkg.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, 0, fmt.Errorf("failed to decode posts: %w", err)
	}

	return posts, total, nil
}

// GetPostsByAuthor retrieves posts by a specific author
func (r *PostRepository) GetPostsByAuthor(ctx context.Context, authorID primitive.ObjectID, pagination postpkg.PostPagination) ([]postpkg.Post, int64, error) {
	filter := postpkg.PostFilter{AuthorID: authorID.Hex()}
	return r.GetPosts(ctx, filter, pagination)
}

// GetPostsByCategory retrieves posts by category
func (r *PostRepository) GetPostsByCategory(ctx context.Context, category string, pagination postpkg.PostPagination) ([]postpkg.Post, int64, error) {
	filter := postpkg.PostFilter{Category: category}
	return r.GetPosts(ctx, filter, pagination)
}

// GetPostsByTag retrieves posts by tag
func (r *PostRepository) GetPostsByTag(ctx context.Context, tag string, pagination postpkg.PostPagination) ([]postpkg.Post, int64, error) {
	filter := postpkg.PostFilter{Tag: tag}
	return r.GetPosts(ctx, filter, pagination)
}

// LikePost adds a user to the post's liked by list
func (r *PostRepository) LikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$addToSet": bson.M{"likedBy": userID},
		"$inc":      bson.M{"likesCount": 1},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to like post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// UnlikePost removes a user from the post's liked by list
func (r *PostRepository) UnlikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$pull": bson.M{"likedBy": userID},
		"$inc":  bson.M{"likesCount": -1},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to unlike post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// IsPostLikedByUser checks if a user has liked a post
func (r *PostRepository) IsPostLikedByUser(ctx context.Context, postID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{
		"_id":     postID,
		"status":  postpkg.PostStatusActive,
		"likedBy": userID,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check if post is liked: %w", err)
	}

	return count > 0, nil
}

// IncrementViewCount increments the view count of a post
func (r *PostRepository) IncrementViewCount(ctx context.Context, postID primitive.ObjectID) error {
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$inc": bson.M{"viewsCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

// UpdateCommentsCount updates the comments count of a post
func (r *PostRepository) UpdateCommentsCount(ctx context.Context, postID primitive.ObjectID, increment int) error {
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$inc": bson.M{"commentsCount": increment},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update comments count: %w", err)
	}

	return nil
}

// SearchPosts searches posts by text query
func (r *PostRepository) SearchPosts(ctx context.Context, query string, filter postpkg.PostFilter, pagination postpkg.PostPagination) ([]postpkg.Post, int64, error) {
	// Build text search filter
	mongoFilter := bson.M{
		"status": postpkg.PostStatusActive,
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"content": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{query}}},
		},
	}

	// Add additional filters
	if filter.Category != "" {
		mongoFilter["category"] = filter.Category
	}
	if filter.AuthorID != "" {
		authorID, err := primitive.ObjectIDFromHex(filter.AuthorID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid author ID: %w", err)
		}
		mongoFilter["authorId"] = authorID
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build find options
	skip := int64((pagination.Page - 1) * pagination.PageSize)
	limit := int64(pagination.PageSize)

	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search posts: %w", err)
	}
	defer cursor.Close(ctx)

	var posts []postpkg.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, 0, fmt.Errorf("failed to decode search results: %w", err)
	}

	return posts, total, nil
}

// GetPopularPosts retrieves popular posts based on engagement metrics
func (r *PostRepository) GetPopularPosts(ctx context.Context, limit int, timeframe string) ([]postpkg.Post, error) {
	// Calculate time filter based on timeframe
	var timeFilter bson.M
	now := time.Now()

	switch timeframe {
	case "day":
		timeFilter = bson.M{"createdAt": bson.M{"$gte": now.AddDate(0, 0, -1)}}
	case "week":
		timeFilter = bson.M{"createdAt": bson.M{"$gte": now.AddDate(0, 0, -7)}}
	case "month":
		timeFilter = bson.M{"createdAt": bson.M{"$gte": now.AddDate(0, -1, 0)}}
	default:
		timeFilter = bson.M{} // All time
	}

	// Build filter
	mongoFilter := bson.M{"status": postpkg.PostStatusActive}
	if len(timeFilter) > 0 {
		for k, v := range timeFilter {
			mongoFilter[k] = v
		}
	}

	// Build find options with popularity sorting
	findOptions := options.Find().
		SetSort(bson.D{
			{Key: "likesCount", Value: -1},
			{Key: "commentsCount", Value: -1},
			{Key: "viewsCount", Value: -1},
		}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular posts: %w", err)
	}
	defer cursor.Close(ctx)

	var posts []postpkg.Post
	if err = cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("failed to decode popular posts: %w", err)
	}

	return posts, nil
}

// GetTrendingTags retrieves trending tags based on recent post activity
func (r *PostRepository) GetTrendingTags(ctx context.Context, limit int) ([]string, error) {
	// Get posts from the last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	pipeline := []bson.M{
		{"$match": bson.M{
			"status":    postpkg.PostStatusActive,
			"createdAt": bson.M{"$gte": sevenDaysAgo},
			"tags":      bson.M{"$exists": true, "$ne": []interface{}{}},
		}},
		{"$unwind": "$tags"},
		{"$group": bson.M{
			"_id":   "$tags",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": limit},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending tags: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode trending tags: %w", err)
	}

	var tags []string
	for _, result := range results {
		if tag, ok := result["_id"].(string); ok {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

// GetPostStats retrieves detailed statistics for a post
func (r *PostRepository) GetPostStats(ctx context.Context, postID primitive.ObjectID) (*postpkg.PostStats, error) {
	var post postpkg.Post
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}

	err := r.collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post stats: %w", err)
	}

	stats := &postpkg.PostStats{
		PostID:        post.ID,
		ViewsCount:    post.ViewsCount,
		LikesCount:    post.LikesCount,
		CommentsCount: post.CommentsCount,
		SharesCount:   0, // TODO: Implement shares tracking
		CreatedAt:     post.CreatedAt.Format("2006-01-02"),
		Category:      post.Category,
		Tags:          post.Tags,
	}

	return stats, nil
}

// ReportPost marks a post as reported
func (r *PostRepository) ReportPost(ctx context.Context, postID primitive.ObjectID) error {
	filter := bson.M{"_id": postID, "status": postpkg.PostStatusActive}
	update := bson.M{
		"$set": bson.M{
			"isReported": true,
			"updatedAt":  time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to report post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// HidePost hides a post from public view
func (r *PostRepository) HidePost(ctx context.Context, postID primitive.ObjectID) error {
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$set": bson.M{
			"status":    postpkg.PostStatusHidden,
			"isHidden":  true,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to hide post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// UnhidePost unhides a previously hidden post
func (r *PostRepository) UnhidePost(ctx context.Context, postID primitive.ObjectID) error {
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$set": bson.M{
			"status":    postpkg.PostStatusActive,
			"isHidden":  false,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to unhide post: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
