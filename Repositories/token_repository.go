package repositories

import (
	"context"

	tokenpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TokenRepository struct {
	collection *mongo.Collection
}

func NewTokenRepository(c *mongo.Collection) *TokenRepository {
	return &TokenRepository{collection: c}
}

func (r *TokenRepository) StoreToken(ctx context.Context, token tokenpkg.Token) error {
	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *TokenRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	filter := bson.M{"refresh_token": refreshToken}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *TokenRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (tokenpkg.Token, error) {
	var token tokenpkg.Token
	filter := bson.M{"refresh_token": refreshToken}
	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return tokenpkg.Token{}, err
	}
	return token, nil
}

func (r *TokenRepository) DeleteTokensByUserID(ctx context.Context, userID string) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"user_id": objID}
	_, err = r.collection.DeleteMany(ctx, filter)
	return err
}
