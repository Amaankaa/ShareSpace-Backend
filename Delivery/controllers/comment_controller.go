package controllers

import (
    "net/http"
    "strconv"

    commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentController struct {
    uc commentpkg.ICommentUsecase
}

func NewCommentController(uc commentpkg.ICommentUsecase) *CommentController { return &CommentController{uc: uc} }

func (cc *CommentController) CreateComment(c *gin.Context) {
    userIDStr := c.GetString("userID")
    if userIDStr == "" { userIDStr = c.GetString("user_id") }
    if userIDStr == "" { c.JSON(http.StatusUnauthorized, gin.H{"error":"User not authenticated"}); return }

    postIDHex := c.Param("id")
    postID, err := primitive.ObjectIDFromHex(postIDHex)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid post ID"}); return }

    var req commentpkg.CreateCommentRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request format"}); return }

    uid, _ := primitive.ObjectIDFromHex(userIDStr)
    resp, err := cc.uc.CreateComment(c.Request.Context(), postID, req, uid)
    if err != nil {
        switch {
        case contains(err.Error(), "not found"):
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        case contains(err.Error(), "unauthorized"):
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        }
        return
    }
    c.JSON(http.StatusCreated, gin.H{"message":"Comment created successfully","comment": resp})
}

func (cc *CommentController) GetComments(c *gin.Context) {
    postIDHex := c.Param("id")
    postID, err := primitive.ObjectIDFromHex(postIDHex)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid post ID"}); return }

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
    sort := c.DefaultQuery("sort", "desc")

    resp, err := cc.uc.GetComments(c.Request.Context(), postID, commentpkg.CommentPagination{Page: page, PageSize: pageSize, Sort: sort})
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }

    c.JSON(http.StatusOK, resp)
}

func (cc *CommentController) DeleteComment(c *gin.Context) {
    userIDStr := c.GetString("userID")
    if userIDStr == "" { userIDStr = c.GetString("user_id") }
    if userIDStr == "" { c.JSON(http.StatusUnauthorized, gin.H{"error":"User not authenticated"}); return }

    commentIDHex := c.Param("commentId")
    commentID, err := primitive.ObjectIDFromHex(commentIDHex)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid comment ID"}); return }

    uid, _ := primitive.ObjectIDFromHex(userIDStr)
    if err := cc.uc.DeleteComment(c.Request.Context(), commentID, uid); err != nil {
        switch {
        case contains(err.Error(), "not found"):
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        case contains(err.Error(), "unauthorized"):
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"message":"Comment deleted successfully"})
}

func (cc *CommentController) UpdateComment(c *gin.Context) {
    userIDStr := c.GetString("userID"); if userIDStr == "" { userIDStr = c.GetString("user_id") }
    if userIDStr == "" { c.JSON(http.StatusUnauthorized, gin.H{"error":"User not authenticated"}); return }

    commentIDHex := c.Param("commentId")
    commentID, err := primitive.ObjectIDFromHex(commentIDHex)
    if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid comment ID"}); return }

    var req commentpkg.UpdateCommentRequest
    if err := c.ShouldBindJSON(&req); err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request format"}); return }
    uid, _ := primitive.ObjectIDFromHex(userIDStr)
    resp, err := cc.uc.UpdateComment(c.Request.Context(), commentID, req, uid)
    if err != nil {
        switch {
        case contains(err.Error(), "not found"):
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        case contains(err.Error(), "unauthorized"):
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        }
        return
    }
    c.JSON(http.StatusOK, gin.H{"message":"Comment updated successfully","comment": resp})
}

// contains is a tiny helper to avoid importing strings in the controller
func contains(s, substr string) bool {
    for i := 0; i+len(substr) <= len(s); i++ {
        if s[i:i+len(substr)] == substr { return true }
    }
    return false
}
