package post

import (
	"context"
	"errors"
	"fmt"
	"reportit-api/internal/config"
	"reportit-api/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	col *mongo.Collection
}

func NewPostRepo(db *mongo.Database) *Repo {
	return &Repo{
		col: db.Collection("reportit_posts"),
	}
}

func (r *Repo) Create(ctx context.Context, post Post) (Post, error) {
	res, err := r.col.InsertOne(ctx, post)
	if err != nil {
		return Post{}, fmt.Errorf("create post failed (%w)", err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return Post{}, fmt.Errorf("create post failed and inserted id is not a objectID")
	}

	post.ID = id

	return post, nil
}

func (r *Repo) FetchAllPosts(ctx context.Context, userID string, page int64, limit int64) ([]Post, int64, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Post{}, 0, err
		}
		return []Post{}, 0, fmt.Errorf("invalid post id (%w)", err)
	}

	filter := bson.M{
		"postedBy": objectID,
	}

	totalPosts, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return []Post{}, 0, err
	}

	skip := (page - 1) * limit

	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return []Post{}, 0, fmt.Errorf("error fetching the posts (%w)", err)
	}

	defer cursor.Close(ctx)

	var posts []Post

	if err := cursor.All(ctx, &posts); err != nil {
		return []Post{}, 0, fmt.Errorf("error decoding the posts (%w)", err)
	}

	return posts, totalPosts, nil
}

func (r *Repo) Delete(ctx context.Context, postID string, userID string, config config.Config) (Post, error) {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Post{}, err
		}
		return Post{}, fmt.Errorf("invalid post id (%w)", err)
	}

	existingPost, err := r.FindByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	if existingPost.PostedByID.Hex() != userID {
		return Post{}, errors.New("unauthorized request for deleting post")
	}

	filter := bson.M{
		"_id": postObjectID,
	}

	var filenames []string
	var tempFilename string

	if len(existingPost.EditHistory) != 0 || existingPost.ImageURL != "" {
		tempFilename = storage.FilenameFromURL(existingPost.ImageURL, config.SupabaseBucket)
		filenames = append(filenames, tempFilename)

		for _, filename := range existingPost.EditHistory {
			tempFilename = storage.FilenameFromURL(filename.ImageURL, config.SupabaseBucket)
			filenames = append(filenames, tempFilename)
		}

		for _, filename := range filenames {
			if err := storage.DeleteImageFromSupabase(filename, config); err != nil {
				return Post{}, fmt.Errorf("unable to delete the image from the database (%w)", err)
			}
		}
	}

	var post Post
	err = r.col.FindOneAndDelete(
		ctx,
		filter,
		options.FindOneAndDelete()).Decode(&post)

	if err != nil {
		return Post{}, fmt.Errorf("%w", err)
	}

	return post, nil
}

func (r *Repo) FindByID(ctx context.Context, postID string) (Post, error) {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Post{}, err
		}
		return Post{}, fmt.Errorf("invalid post id (%w)", err)
	}

	filter := bson.M{
		"_id": postObjectID,
	}
	var post Post
	err = r.col.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Post{}, fmt.Errorf("post not found (%w)", err)
		}
		return Post{}, fmt.Errorf("%w", err)
	}

	return post, nil
}

func (r *Repo) Update(ctx context.Context, postID string, post UpdatePost) (Post, error) {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Post{}, err
		}
		return Post{}, fmt.Errorf("invalid post id (%w)", err)
	}

	filter := bson.M{
		"_id": postObjectID,
	}

	existingPost, err := r.FindByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	edit := UpdatePost{
		Title:       existingPost.Title,
		Description: existingPost.Description,
		PostType:    existingPost.PostType,
		ImageURL:    existingPost.ImageURL,
		UpdatedAt:   existingPost.UpdatedAt,
	}

	update := bson.M{
		"$set": post,
		"$push": bson.M{
			"editHistory": edit,
		},
	}

	var updatedPost Post

	err = r.col.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedPost)

	if err != nil {
		return Post{}, fmt.Errorf("updating details failed (%w)", err)
	}

	return updatedPost, nil
}
