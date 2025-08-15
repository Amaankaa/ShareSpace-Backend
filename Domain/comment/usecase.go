package comment

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockery --name=ICommentUsecase --output=../../mocks --outpkg=mocks

type ICommentUsecase interface {
	CreateComment(ctx context.Context, postID primitive.ObjectID, req CreateCommentRequest, userID primitive.ObjectID) (*CommentResponse, error)
	GetComments(ctx context.Context, postID primitive.ObjectID, pagination CommentPagination) (*CommentListResponse, error)
	UpdateComment(ctx context.Context, commentID primitive.ObjectID, req UpdateCommentRequest, userID primitive.ObjectID) (*CommentResponse, error)
	DeleteComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error
}
