package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *Repo {
	return &Repo{
		col: db.Collection("reportit_users"),
	}
}

func (r *Repo) FindByEmail(ctx context.Context, email string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// creating filter
	filter := bson.M{
		"email": email,
	}

	var user User
	err := r.col.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, mongo.ErrNoDocuments
		}
		return User{}, fmt.Errorf("find by email failed (%w)", err)
	}

	return user, nil
}

func (r *Repo) Create(ctx context.Context, user User) (User, error) {
	res, err := r.col.InsertOne(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("insert user failed (%w)", err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return User{}, fmt.Errorf("insert user failed and inserted id is not a objectID")
	}
	user.ID = id

	return user, nil
}

func (r *Repo) Update(ctx context.Context, user UpdateUser, userID string) (User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, mongo.ErrNoDocuments
		}
		return User{}, fmt.Errorf("invalid id provided (%w)", err)
	}

	filter := bson.M{
		"_id": objectID,
	}
	update := bson.M{
		"$set": user,
	}
	var updatedUser User

	err = r.col.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedUser)

	if err != nil {
		return User{}, fmt.Errorf("updating details failed (%w)", err)
	}
	return updatedUser, nil
}

func (r *Repo) FetchAllUser(ctx context.Context, page int64, limit int64) ([]User, int64, error) {
	totalUser, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return []User{}, 0, fmt.Errorf("error counting users from database (%w)", err)
	}

	skip := (page - 1) * limit

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []User{}, 0, fmt.Errorf("error fetching the users (%w)", err)
	}

	defer cursor.Close(ctx)

	var users []User

	if err := cursor.All(ctx, &users); err != nil {
		return []User{}, 0, fmt.Errorf("error decoding the users (%w)", err)
	}

	return users, totalUser, nil
}

func (r *Repo) FetchSingleUser(ctx context.Context, userID string) (User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, mongo.ErrNoDocuments
		}
		return User{}, fmt.Errorf("invalid id provided (%w)", err)
	}

	filter := bson.M{
		"_id": objectID,
	}

	var user User
	err = r.col.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		return User{}, fmt.Errorf("failed to get the user details (%w)", err)
	}
	return user, nil
}

func (r *Repo) Delete(ctx context.Context, userID string) (User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return User{}, err
		}
		return User{}, fmt.Errorf("invalid user id (%w)", err)
	}

	filter := bson.M{
		"_id": objectID,
	}

	var user User

	err = r.col.FindOneAndDelete(
		ctx,
		filter,
		options.FindOneAndDelete()).Decode(&user)

	if err != nil {
		return User{}, fmt.Errorf("%w", err)
	}

	return user, nil
}
