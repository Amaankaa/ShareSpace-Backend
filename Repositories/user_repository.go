package repositories

import (
	"context"
	"errors"
	"reflect"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: collection,
	}
}

// Check if username exists
func (ur *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var user userpkg.User
	err := ur.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Check if email exists
func (ur *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var user userpkg.User
	err := ur.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Count how many users exist (used to decide admin role)
func (ur *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	count, err := ur.collection.CountDocuments(ctx, bson.M{})
	return count, err
}

// Save new user to DB
func (ur *UserRepository) CreateUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	user.ID = primitive.NewObjectID()
	user.IsVerified = false

	_, err := ur.collection.InsertOne(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	user.Password = "" // don’t return hashed password
	return user, nil
}

func (ur *UserRepository) GetUserByLogin(ctx context.Context, login string) (userpkg.User, error) {
	var user userpkg.User
	filter := bson.M{
		"$or": []bson.M{
			{"username": login},
			{"email": login},
		},
	}
	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return userpkg.User{}, errors.New("user not found")
	}
	return user, err
}

func (ur *UserRepository) FindByID(ctx context.Context, userID string) (userpkg.User, error) {
	var user userpkg.User

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return userpkg.User{}, err
	}

	err = ur.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		return userpkg.User{}, err
	}

	return user, nil
}

func (ur *UserRepository) UpdateUserRoleByID(ctx context.Context, userID, role string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": bson.M{"role": role}}
	res, err := ur.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateRoleAndPromoter atomically updates role and promoted_by in a single write.
func (ur *UserRepository) UpdateRoleAndPromoter(ctx context.Context, userID string, role string, promoterID *string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid}
	set := bson.M{"role": role}
	update := bson.M{"$set": set}
	if promoterID != nil {
		// set promoted_by
		promoterOID, err := primitive.ObjectIDFromHex(*promoterID)
		if err != nil {
			return err
		}
		set["promoted_by"] = promoterOID
	} else {
		// unset promoted_by
		update["$unset"] = bson.M{"promoted_by": ""}
	}
	res, err := ur.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateIsVerifiedByEmail sets a user's verification status by email
func (ur *UserRepository) UpdateIsVerifiedByEmail(ctx context.Context, email string, verified bool) error {
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"isVerified": verified}}
	res, err := ur.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (ur *UserRepository) UpdateProfile(ctx context.Context, userID string, updates userpkg.UpdateProfileRequest) (userpkg.User, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return userpkg.User{}, err
    }
    updateDoc := bson.M{
        "$set": bson.M{
            "updatedAt": time.Now(),
        },
    }
    if updates.Fullname != "" {
        updateDoc["$set"].(bson.M)["fullname"] = updates.Fullname
    }
    if updates.Bio != "" {
        updateDoc["$set"].(bson.M)["bio"] = updates.Bio
    }
    if updates.ProfilePicture != "" {
        updateDoc["$set"].(bson.M)["profilePicture"] = updates.ProfilePicture
    }
	// Only update contactInfo if it is not empty
    if !reflect.DeepEqual(updates.ContactInfo, userpkg.ContactInfo{}) {
        updateDoc["$set"].(bson.M)["contactInfo"] = updates.ContactInfo
    }
	filter := bson.M{"_id": oid}
    _, err = ur.collection.UpdateOne(ctx, filter, updateDoc)
    if err != nil {
        return userpkg.User{}, err
    }
    return ur.FindByID(ctx, userID)
}

func (ur *UserRepository) GetUserProfile(ctx context.Context, userID string) (userpkg.User, error) {
    user, err := ur.FindByID(ctx, userID)
    if err != nil {
        return userpkg.User{}, err
    }
    user.Password = "" // Don't return password
    return user, nil
}
