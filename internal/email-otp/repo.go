package emailotp

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repo struct {
	col *mongo.Collection
}

func NewOTPRepo(db *mongo.Database) *Repo {
	return &Repo{
		col: db.Collection("reportit_otp"),
	}
}

func (r *Repo) DeleteByEmail(ctx context.Context, email string) error {
	filter := bson.M{
		"email": email,
	}

	_, err := r.col.DeleteMany(ctx, filter)
	return err
}

func (r *Repo) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id (%w)", err)
	}

	filter := bson.M{
		"_id": objectID,
	}

	_, err = r.col.DeleteOne(ctx, filter)
	return err
}

func (r *Repo) Create(ctx context.Context, otp EmailOTP) error {
	_, err := r.col.InsertOne(ctx, otp)
	return err
}

func (r *Repo) FindByEmail(ctx context.Context, email string) (*EmailOTP, error) {
	filter := bson.M{
		"email": email,
	}

	var emailOTP EmailOTP

	err := r.col.FindOne(ctx, filter).Decode(&emailOTP)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &EmailOTP{}, mongo.ErrNoDocuments
		}
		return &EmailOTP{}, err
	}

	return &emailOTP, nil
}
