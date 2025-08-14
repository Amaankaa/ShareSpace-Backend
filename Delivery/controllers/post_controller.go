package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostController struct {
	postUsecase postpkg.PostUsecase
}

func NewPostController(postUsecase postpkg.PostUsecase) *PostController {
	return &PostController{
		postUsecase: postUsecase,
	}
}

// CreatePost handles POST /posts
func (ctrl *PostController) CreatePost(c *gin.Context) {
	var req postpkg.CreatePostRequest

	// Parse JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Get user ID from JWT token (set by auth middleware)
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Create post
	post, err := ctrl.postUsecase.CreatePost(ctx, req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post created successfully",
		"post":    post,
	})
}

// GetPosts handles GET /posts
func (ctrl *PostController) GetPosts(c *gin.Context) {
	// Parse query parameters
	filter := postpkg.PostFilter{
		Category: c.Query("category"),
		AuthorID: c.Query("authorId"),
		Tag:      c.Query("tag"),
	}

	// Parse year filter
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.Year = year
		}
	}

	// Parse anonymous filter
	if anonStr := c.Query("isAnonymous"); anonStr != "" {
		if anon, err := strconv.ParseBool(anonStr); err == nil {
			filter.IsAnonymous = &anon
		}
	}

	// Parse pagination
	pagination := postpkg.PostPagination{
		Page:      1,
		PageSize:  20,
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pagination.Page = page
		}
	}

	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			pagination.PageSize = size
		}
	}

	// Get viewer ID for like status (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get posts
	result, err := ctrl.postUsecase.GetPosts(ctx, filter, pagination, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPost handles GET /posts/:id
func (ctrl *PostController) GetPost(c *gin.Context) {
	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get viewer ID for like status (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get post
	post, err := ctrl.postUsecase.GetPost(ctx, postID, viewerID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// UpdatePost handles PATCH /posts/:id
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req postpkg.UpdatePostRequest

	// Parse JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Get user ID from JWT token
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Update post
	post, err := ctrl.postUsecase.UpdatePost(ctx, postID, req, userID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		if err.Error() == "unauthorized: only the author can update this post" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own posts"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    post,
	})
}

// DeletePost handles DELETE /posts/:id
func (ctrl *PostController) DeletePost(c *gin.Context) {
	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get user ID from JWT token
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Delete post
	err = ctrl.postUsecase.DeletePost(ctx, postID, userID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		if err.Error() == "unauthorized: only the author can delete this post" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own posts"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
	})
}

// LikePost handles POST /posts/:id/like
func (ctrl *PostController) LikePost(c *gin.Context) {
	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get user ID from JWT token
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Like post
	err = ctrl.postUsecase.LikePost(ctx, postID, userID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		if err.Error() == "post already liked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Post already liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post liked successfully",
	})
}

// UnlikePost handles DELETE /posts/:id/like
func (ctrl *PostController) UnlikePost(c *gin.Context) {
	// Parse post ID
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get user ID from JWT token
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Unlike post
	err = ctrl.postUsecase.UnlikePost(ctx, postID, userID)
	if err != nil {
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		if err.Error() == "post not liked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Post not liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post unliked successfully",
	})
}

// SearchPosts handles GET /posts/search
func (ctrl *PostController) SearchPosts(c *gin.Context) {
	// Get search query
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Parse filters
	filter := postpkg.PostFilter{
		Category: c.Query("category"),
		AuthorID: c.Query("authorId"),
	}

	// Parse pagination
	pagination := postpkg.PostPagination{
		Page:      1,
		PageSize:  20,
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pagination.Page = page
		}
	}

	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			pagination.PageSize = size
		}
	}

	// Get viewer ID (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Search posts
	result, err := ctrl.postUsecase.SearchPosts(ctx, query, filter, pagination, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPopularPosts handles GET /posts/popular
func (ctrl *PostController) GetPopularPosts(c *gin.Context) {
	// Parse parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	timeframe := c.DefaultQuery("timeframe", "week") // day, week, month, all

	// Get viewer ID (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get popular posts
	result, err := ctrl.postUsecase.GetPopularPosts(ctx, limit, timeframe, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTrendingTags handles GET /posts/trending-tags
func (ctrl *PostController) GetTrendingTags(c *gin.Context) {
	// Parse limit
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get trending tags
	tags, err := ctrl.postUsecase.GetTrendingTags(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

// GetUserPosts handles GET /users/:userId/posts
func (ctrl *PostController) GetUserPosts(c *gin.Context) {
	// Parse user ID
	userID, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse pagination
	pagination := postpkg.PostPagination{
		Page:      1,
		PageSize:  20,
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pagination.Page = page
		}
	}

	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			pagination.PageSize = size
		}
	}

	// Get viewer ID (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get user posts
	result, err := ctrl.postUsecase.GetUserPosts(ctx, userID, pagination, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPostsByCategory handles GET /posts/category/:category
func (ctrl *PostController) GetPostsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	// Parse pagination
	pagination := postpkg.PostPagination{
		Page:      1,
		PageSize:  20,
		SortBy:    c.DefaultQuery("sortBy", "createdAt"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pagination.Page = page
		}
	}

	if sizeStr := c.Query("pageSize"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			pagination.PageSize = size
		}
	}

	// Get viewer ID (optional)
	var viewerID *primitive.ObjectID
	if userIDStr, exists := c.Get("userID"); exists {
		if userObjID, err := primitive.ObjectIDFromHex(userIDStr.(string)); err == nil {
			viewerID = &userObjID
		}
	}

	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get posts by category
	result, err := ctrl.postUsecase.GetPostsByCategory(ctx, category, pagination, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
