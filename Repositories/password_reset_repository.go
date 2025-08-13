package repositories

import (
    "context"

    userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type PasswordResetRepo struct {
    passwordResetCollection *mongo.Collection
    userCollection *mongo.Collection
}

func NewPasswordResetRepo(r *mongo.Collection, uc *mongo.Collection) *PasswordResetRepo {
    return &PasswordResetRepo{
        passwordResetCollection: r,
        userCollection: uc,
    }
}

func (r *PasswordResetRepo) StoreResetRequest(ctx context.Context, reset userpkg.PasswordReset) error {
    _, err := r.passwordResetCollection.InsertOne(ctx, reset)
    return err
}

func (r *PasswordResetRepo) GetResetRequest(ctx context.Context, email string) (userpkg.PasswordReset, error) {
    var reset userpkg.PasswordReset

    err := r.passwordResetCollection.FindOne(ctx, bson.M{"email": email}).Decode(&reset)
    return reset, err
}

func (r *PasswordResetRepo) DeleteResetRequest(ctx context.Context, email string) error {
    _, err := r.passwordResetCollection.DeleteOne(ctx, bson.M{"email": email})
    return err
}

func (r *PasswordResetRepo) IncrementAttemptCount(ctx context.Context, email string) error {
    _, err := r.passwordResetCollection.UpdateOne(
        ctx,
        bson.M{"email": email},
        bson.M{"$inc": bson.M{"attemptcount": 1}},
    )
    return err
}

func (r *UserRepository) UpdatePasswordByEmail(ctx context.Context, email, hashedPassword string) error {
    filter := bson.M{"email": email}
    update := bson.M{"$set": bson.M{"password": hashedPassword}}
    _, err := r.collection.UpdateOne(ctx, filter, update)
    return err
}