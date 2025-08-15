package repositories

import (
    "context"
    "fmt"
    "time"

    commentpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/comment"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type CommentRepository struct {
    collection *mongo.Collection
}

func NewCommentRepository(collection *mongo.Collection) *CommentRepository {
    return &CommentRepository{collection: collection}
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment commentpkg.Comment) (*commentpkg.Comment, error) {
    comment.ID = primitive.NewObjectID()
    now := time.Now()
    comment.CreatedAt = now
    comment.UpdatedAt = now

    if _, err := r.collection.InsertOne(ctx, comment); err != nil {
        return nil, fmt.Errorf("failed to create comment: %w", err)
    }
    return &comment, nil
}

func (r *CommentRepository) GetCommentsByPost(ctx context.Context, postID primitive.ObjectID, pagination commentpkg.CommentPagination) ([]commentpkg.Comment, int64, error) {
    filter := bson.M{"postId": postID}

    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count comments: %w", err)
    }

    sortOrder := -1
    if pagination.Sort == "asc" {
        sortOrder = 1
    }

    skip := int64((pagination.Page - 1) * pagination.PageSize)
    limit := int64(pagination.PageSize)

    opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: sortOrder}}).SetSkip(skip).SetLimit(limit)

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to find comments: %w", err)
    }
    defer cursor.Close(ctx)

    var comments []commentpkg.Comment
    if err := cursor.All(ctx, &comments); err != nil {
        return nil, 0, fmt.Errorf("failed to decode comments: %w", err)
    }

    return comments, total, nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*commentpkg.Comment, error) {
    var cmt commentpkg.Comment
    if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&cmt); err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("comment not found")
        }
        return nil, fmt.Errorf("failed to get comment: %w", err)
    }
    return &cmt, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, id primitive.ObjectID) error {
    res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return fmt.Errorf("failed to delete comment: %w", err)
    }
    if res.DeletedCount == 0 {
        return fmt.Errorf("comment not found")
    }
    return nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, id primitive.ObjectID, content string) (*commentpkg.Comment, error) {
    update := bson.M{"$set": bson.M{"content": content, "updatedAt": time.Now()}}
    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    var updated commentpkg.Comment
    if err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&updated); err != nil {
        if err == mongo.ErrNoDocuments { return nil, fmt.Errorf("comment not found") }
        return nil, fmt.Errorf("failed to update comment: %w", err)
    }
    return &updated, nil
}
