package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	resourcepkg "github.com/Amaankaa/Blog-Starter-Project/Domain/resource"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceController struct {
	usecase resourcepkg.ResourceUsecase
}

func NewResourceController(usecase resourcepkg.ResourceUsecase) *ResourceController {
	return &ResourceController{usecase: usecase}
}

// POST /resources
func (ctrl *ResourceController) CreateResource(c *gin.Context) {
	var req resourcepkg.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.CreateResource(ctx, req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Resource created successfully", "resource": res})
}

// GET /resources
func (ctrl *ResourceController) GetResources(c *gin.Context) {
	filter := resourcepkg.ResourceFilter{Type: c.Query("type"), Category: c.Query("category"), CreatorID: c.Query("creatorId"), Tag: c.Query("tag"), Difficulty: c.Query("difficulty")}
	if v := c.Query("isVerified"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filter.IsVerified = &b
		}
	}
	if v := c.Query("hasDeadline"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filter.HasDeadline = &b
		}
	}
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20, SortBy: c.DefaultQuery("sortBy", "createdAt"), SortOrder: c.DefaultQuery("sortOrder", "desc")}
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pg.Page = n
		}
	}
	if s := c.Query("pageSize"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			pg.PageSize = n
		}
	}
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetResources(ctx, filter, pg, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /resources/:id
func (ctrl *ResourceController) GetResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetResource(ctx, id, viewerID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// PATCH /resources/:id
func (ctrl *ResourceController) UpdateResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var req resourcepkg.UpdateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.UpdateResource(ctx, id, req, userID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "unauthorized: only the creator can update this resource" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own resources"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource updated successfully", "resource": res})
}

// DELETE /resources/:id
func (ctrl *ResourceController) DeleteResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ctrl.usecase.DeleteResource(ctx, id, userID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "unauthorized: only the creator can delete this resource" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own resources"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

// POST /resources/:id/like
func (ctrl *ResourceController) LikeResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ctrl.usecase.LikeResource(ctx, id, userID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "resource already liked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Resource already liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource liked successfully"})
}

// DELETE /resources/:id/like
func (ctrl *ResourceController) UnlikeResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ctrl.usecase.UnlikeResource(ctx, id, userID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "resource not liked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Resource not liked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource unliked successfully"})
}

// POST /resources/:id/bookmark
func (ctrl *ResourceController) BookmarkResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ctrl.usecase.BookmarkResource(ctx, id, userID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "resource already bookmarked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Resource already bookmarked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource bookmarked successfully"})
}

// DELETE /resources/:id/bookmark
func (ctrl *ResourceController) UnbookmarkResource(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ctrl.usecase.UnbookmarkResource(ctx, id, userID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "resource not bookmarked by user" {
			c.JSON(http.StatusConflict, gin.H{"error": "Resource not bookmarked"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource unbookmarked successfully"})
}

// GET /resources/search
func (ctrl *ResourceController) SearchResources(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}
	filter := resourcepkg.ResourceFilter{Category: c.Query("category"), Type: c.Query("type")}
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20, SortBy: c.DefaultQuery("sortBy", "createdAt"), SortOrder: c.DefaultQuery("sortOrder", "desc")}
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pg.Page = n
		}
	}
	if s := c.Query("pageSize"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			pg.PageSize = n
		}
	}
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.SearchResources(ctx, query, filter, pg, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /resources/popular, /trending, /top-rated
func (ctrl *ResourceController) GetPopularResources(c *gin.Context) {
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	timeframe := c.DefaultQuery("timeframe", "week")
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetPopularResources(ctx, limit, timeframe, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *ResourceController) GetTrendingResources(c *gin.Context) {
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetTrendingResources(ctx, limit, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *ResourceController) GetTopRatedResources(c *gin.Context) {
	limit := 20
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	category := c.Query("category")
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetTopRatedResources(ctx, limit, category, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /users/:id/resources, /liked, /bookmarked
func (ctrl *ResourceController) GetUserResources(c *gin.Context) {
	uid, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20, SortBy: c.DefaultQuery("sortBy", "createdAt"), SortOrder: c.DefaultQuery("sortOrder", "desc")}
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pg.Page = n
		}
	}
	if s := c.Query("pageSize"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			pg.PageSize = n
		}
	}
	var viewerID *primitive.ObjectID
	if v, ok := c.Get("userID"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	} else if v, ok := c.Get("user_id"); ok {
		if id, err := primitive.ObjectIDFromHex(v.(string)); err == nil {
			viewerID = &id
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetUserResources(ctx, uid, pg, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *ResourceController) GetUserLikedResources(c *gin.Context) {
	uid, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20, SortBy: c.DefaultQuery("sortBy", "createdAt"), SortOrder: c.DefaultQuery("sortOrder", "desc")}
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pg.Page = n
		}
	}
	if s := c.Query("pageSize"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			pg.PageSize = n
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetUserLikedResources(ctx, uid, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctrl *ResourceController) GetUserBookmarkedResources(c *gin.Context) {
	uid, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	pg := resourcepkg.ResourcePagination{Page: 1, PageSize: 20, SortBy: c.DefaultQuery("sortBy", "createdAt"), SortOrder: c.DefaultQuery("sortOrder", "desc")}
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			pg.Page = n
		}
	}
	if s := c.Query("pageSize"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			pg.PageSize = n
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	res, err := ctrl.usecase.GetUserBookmarkedResources(ctx, uid, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /resources/:id/analytics (creator-only)
func (ctrl *ResourceController) GetResourceAnalytics(c *gin.Context) {
	resID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	analytics, err := ctrl.usecase.GetResourceAnalytics(ctx, resID, userID)
	if err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		if err.Error() == "unauthorized: only the creator can view analytics" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view analytics for your own resources"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

// GET /users/:id/resources/stats
func (ctrl *ResourceController) GetUserResourceStats(c *gin.Context) {
	uid, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	stats, err := ctrl.usecase.GetUserResourceStats(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

type reportRequest struct {
	Reason string `json:"reason"`
}

// POST /resources/:id/report
func (ctrl *ResourceController) ReportResource(c *gin.Context) {
	resID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req reportRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ctrl.usecase.ReportResource(ctx, resID, userID, req.Reason); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource reported successfully"})
}

// POST /resources/:id/verify (admin-only via router middleware)
func (ctrl *ResourceController) VerifyResource(c *gin.Context) {
	resID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}
	var uidStr string
	if v, ok := c.Get("userID"); ok {
		uidStr = v.(string)
	} else if v, ok := c.Get("user_id"); ok {
		uidStr = v.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	verifierID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ctrl.usecase.VerifyResource(ctx, resID, verifierID); err != nil {
		if err.Error() == "resource not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource verified successfully"})
}
