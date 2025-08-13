package repositories

import (
	"context"
	"errors"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VerificationRepo struct {
	collection *mongo.Collection
}

func NewVerificationRepo(collection *mongo.Collection) *VerificationRepo {
	return &VerificationRepo{collection: collection}
}

// StoreVerification upserts a verification record: replaces existing one or inserts new
func (r *VerificationRepo) StoreVerification(ctx context.Context, v userpkg.Verification) error {
	// build filter and update for upsert
	filter := bson.M{"email": v.Email}
	update := bson.M{"$set": bson.M{
		"otp":          v.OTP,
		"expiresAt":    v.ExpiresAt,
		"attemptCount": 0,
	}}
	_, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *VerificationRepo) GetVerification(ctx context.Context, email string) (userpkg.Verification, error) {
	var v userpkg.Verification
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&v)
	return v, err
}

func (r *VerificationRepo) DeleteVerification(ctx context.Context, email string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"email": email})
	return err
}

// IncrementAttemptCount increments attemptCount and errors if no record or on failure
func (r *VerificationRepo) IncrementAttemptCount(ctx context.Context, email string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		// increment the attemptCount field matching struct tag
		bson.M{"$inc": bson.M{"attemptCount": 1}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("verification record not found for email")
	}
	return nil
}
