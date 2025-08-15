package comment

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a comment on a post
type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID `bson:"postId" json:"postId"`
	AuthorID  primitive.ObjectID `bson:"authorId" json:"authorId"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type CommentPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Sort     string `json:"sort"` // createdAt asc|desc
}

type AuthorInfo struct {
	ID             primitive.ObjectID `json:"id"`
	DisplayName    string             `json:"displayName"`
	ProfilePicture string             `json:"profilePicture"`
}

type CommentResponse struct {
	ID        primitive.ObjectID `json:"id"`
	PostID    primitive.ObjectID `json:"postId"`
	Author    AuthorInfo         `json:"author"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
	HasNext    bool              `json:"hasNext"`
	HasPrev    bool              `json:"hasPrev"`
}
