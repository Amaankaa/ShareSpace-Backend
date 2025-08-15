package comment

import (
    "context"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockery --name=ICommentRepository --output=../../mocks --outpkg=mocks

// ICommentRepository defines data operations for comments
type ICommentRepository interface {
    CreateComment(ctx context.Context, comment Comment) (*Comment, error)
    GetCommentsByPost(ctx context.Context, postID primitive.ObjectID, pagination CommentPagination) ([]Comment, int64, error)
    GetByID(ctx context.Context, id primitive.ObjectID) (*Comment, error)
    UpdateComment(ctx context.Context, id primitive.ObjectID, content string) (*Comment, error)
    DeleteComment(ctx context.Context, id primitive.ObjectID) error
}
