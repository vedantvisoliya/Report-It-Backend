package user

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshRepo struct {
	col *mongo.Collection
}

func NewRefreshRepo(db *mongo.Database) *RefreshRepo {
	return &RefreshRepo{
		col: db.Collection("reportit_refresh_tokens"),
	}
}

func (r *RefreshRepo) Create(ctx context.Context, rt RefreshToken) error {
	_, err := r.col.InsertOne(ctx, rt)
	if err != nil {
		return fmt.Errorf("insert refresh token failed (%w)", err)
	}
	return nil
}

func (r *RefreshRepo) FindByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	filter := bson.M{
		"tokenHash": tokenHash,
	}

	var rt RefreshToken
	err := r.col.FindOne(ctx, filter).Decode(&rt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return RefreshToken{}, mongo.ErrNoDocuments
		}
		return RefreshToken{}, fmt.Errorf("refresh token not found (%w)", err)
	}
	return rt, nil
}

func (r *RefreshRepo) Revoke(ctx context.Context, tokenHash string) error {
	filter := bson.M{
		"tokenHash": tokenHash,
	}
	update := bson.M{
		"$set": bson.M{
			"revoked": true,
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("revoke refresh token failed (%w)", err)
	}

	return nil
}

func (r *RefreshRepo) RevokeAllForUser(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{
		"userID": userID,
	}
	update := bson.M{
		"$set": bson.M{
			"revoked": true,
		},
	}

	_, err := r.col.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("revoke all refresh token failed (%w)", err)
	}

	return nil
}
