package usecases

import (
    "context"
    "errors"
    "fmt"
    "math"
    "strings"

    commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
    postpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/post"
    userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentUsecase struct {
    commentRepo commentpkg.ICommentRepository
    postRepo    postpkg.PostRepository
    userRepo    userpkg.IUserRepository
}

func NewCommentUsecase(commentRepo commentpkg.ICommentRepository, postRepo postpkg.PostRepository, userRepo userpkg.IUserRepository) *CommentUsecase {
    return &CommentUsecase{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo}
}

func (uc *CommentUsecase) CreateComment(ctx context.Context, postID primitive.ObjectID, req commentpkg.CreateCommentRequest, userID primitive.ObjectID) (*commentpkg.CommentResponse, error) {
    content := strings.TrimSpace(req.Content)
    if content == "" {
        return nil, errors.New("comment content cannot be empty")
    }

    // Ensure post exists
    if _, err := uc.postRepo.GetPostByID(ctx, postID); err != nil {
        return nil, err
    }

    comment := commentpkg.Comment{PostID: postID, AuthorID: userID, Content: content}
    created, err := uc.commentRepo.CreateComment(ctx, comment)
    if err != nil {
        return nil, fmt.Errorf("failed to create comment: %w", err)
    }

    // Increment comments count on post (non-critical)
    _ = uc.postRepo.UpdateCommentsCount(ctx, postID, 1)

    // Get author for response
    author, err := uc.userRepo.FindByID(ctx, userID.Hex())
    if err != nil {
        return nil, fmt.Errorf("failed to get author: %w", err)
    }

    return &commentpkg.CommentResponse{
        ID:     created.ID,
        PostID: postID,
        Author: commentpkg.AuthorInfo{ID: author.ID, DisplayName: author.DisplayName, ProfilePicture: author.ProfilePicture},
        Content: created.Content,
        CreatedAt: created.CreatedAt,
        UpdatedAt: created.UpdatedAt,
    }, nil
}

func (uc *CommentUsecase) GetComments(ctx context.Context, postID primitive.ObjectID, pagination commentpkg.CommentPagination) (*commentpkg.CommentListResponse, error) {
    if pagination.Page < 1 { pagination.Page = 1 }
    if pagination.PageSize < 1 || pagination.PageSize > 100 { pagination.PageSize = 20 }

    // Ensure post exists
    if _, err := uc.postRepo.GetPostByID(ctx, postID); err != nil {
        return nil, err
    }

    comments, total, err := uc.commentRepo.GetCommentsByPost(ctx, postID, pagination)
    if err != nil {
        return nil, fmt.Errorf("failed to get comments: %w", err)
    }

    // Enrich with author info
    var responses []commentpkg.CommentResponse
    for _, c := range comments {
        author, err := uc.userRepo.FindByID(ctx, c.AuthorID.Hex())
        if err != nil { return nil, fmt.Errorf("failed to get author: %w", err) }
        responses = append(responses, commentpkg.CommentResponse{
            ID: c.ID,
            PostID: c.PostID,
            Author: commentpkg.AuthorInfo{ID: author.ID, DisplayName: author.DisplayName, ProfilePicture: author.ProfilePicture},
            Content: c.Content,
            CreatedAt: c.CreatedAt,
            UpdatedAt: c.UpdatedAt,
        })
    }

    totalPages := int(math.Ceil(float64(total)/float64(pagination.PageSize)))
    hasNext := pagination.Page < totalPages
    hasPrev := pagination.Page > 1

    return &commentpkg.CommentListResponse{Comments: responses, Total: total, Page: pagination.Page, PageSize: pagination.PageSize, TotalPages: totalPages, HasNext: hasNext, HasPrev: hasPrev}, nil
}

func (uc *CommentUsecase) DeleteComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error {
    cmt, err := uc.commentRepo.GetByID(ctx, commentID)
    if err != nil { return err }

    // Only author can delete
    if cmt.AuthorID != userID {
        return errors.New("unauthorized: only the author can delete this comment")
    }

    if err := uc.commentRepo.DeleteComment(ctx, commentID); err != nil {
        return err
    }

    // decrement count (best effort)
    _ = uc.postRepo.UpdateCommentsCount(ctx, cmt.PostID, -1)
    return nil
}

// Ensure interface compliance
var _ commentpkg.ICommentUsecase = (*CommentUsecase)(nil)

func (uc *CommentUsecase) UpdateComment(ctx context.Context, commentID primitive.ObjectID, req commentpkg.UpdateCommentRequest, userID primitive.ObjectID) (*commentpkg.CommentResponse, error) {
    content := strings.TrimSpace(req.Content)
    if content == "" { return nil, errors.New("comment content cannot be empty") }

    cmt, err := uc.commentRepo.GetByID(ctx, commentID)
    if err != nil { return nil, err }
    if cmt.AuthorID != userID { return nil, errors.New("unauthorized: only the author can update this comment") }

    updated, err := uc.commentRepo.UpdateComment(ctx, commentID, content)
    if err != nil { return nil, err }

    author, err := uc.userRepo.FindByID(ctx, userID.Hex())
    if err != nil { return nil, fmt.Errorf("failed to get author: %w", err) }

    return &commentpkg.CommentResponse{
        ID: updated.ID,
        PostID: updated.PostID,
        Author: commentpkg.AuthorInfo{ID: author.ID, DisplayName: author.DisplayName, ProfilePicture: author.ProfilePicture},
        Content: updated.Content,
        CreatedAt: updated.CreatedAt,
        UpdatedAt: updated.UpdatedAt,
    }, nil
}
