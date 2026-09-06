package post

import (
	"context"
	"errors"
	"fmt"
	"image"
	"math"
	"mime/multipart"
	"reportit-api/internal/config"
	"reportit-api/internal/storage"
	"reportit-api/internal/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	repo   *Repo
	config config.Config
}

type CreatePostForm struct {
	Title       string                `form:"title" binding:"required"`
	Description string                `form:"description" binding:"required"`
	PostType    string                `form:"postType" binding:"required"`
	IsAnonymous bool                  `form:"isAnonymous"`
	Image       *multipart.FileHeader `form:"image"`
}

type DeletePostRequest struct {
	UserID string `json:"userID" binding:"required"`
}

func NewPostService(repo *Repo, config config.Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

func (svc *Service) CreatePost(ctx context.Context, input CreatePostForm, userID string) (Post, error) {
	title := strings.TrimSpace(input.Title)
	descp := strings.TrimSpace(input.Description)
	postType := strings.TrimSpace(input.PostType)

	if title == "" {
		return Post{}, errors.New("title not provided")
	} else if len(title) > 100 {
		return Post{}, errors.New("title length exceeded (must be under 100 characters)")
	}

	if descp == "" {
		return Post{}, errors.New("description not provided")
	} else if len(descp) > 5000 {
		return Post{}, errors.New("description length exceeded (must be under 5000 characters)")
	}

	if postType == "" {
		return Post{}, errors.New("post type not provided")
	}

	var imageURL string

	if input.Image != nil {

		file, _, err := utils.ValidateImageType(input.Image)
		if err != nil {
			return Post{}, err
		}

		img, _, err := image.Decode(file)
		file.Close()

		if err != nil {
			return Post{}, fmt.Errorf("invalid or corrupted image (%w)", err)
		}

		resized := utils.ResizeIfLarge(img, utils.MaxImageWidth)
		compressed, err := utils.CompressJpeg(resized, utils.MaxImageSizeBytes)
		if err != nil {
			return Post{}, fmt.Errorf("compression image failed (%w)", err)
		}

		filename := uuid.New().String() + ".jpg"
		url, err := storage.UploadImageToSupabase(compressed, "image/jpeg", string(filename), svc.config)
		if err != nil {
			return Post{}, fmt.Errorf("image upload failed (%w)", err)
		}

		imageURL = url
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Post{}, err
		}
		return Post{}, fmt.Errorf("invalid user id (%w)", err)
	}

	createPost := Post{
		ID:          primitive.NewObjectID(),
		PostedByID:  objectID,
		Title:       title,
		Description: descp,
		ImageURL:    imageURL,
		PostType:    postType,
		IsAnonymous: input.IsAnonymous,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	post, err := svc.repo.Create(ctx, createPost)
	if err != nil {
		return Post{}, fmt.Errorf("%w", err)
	}

	return post, nil
}

func (svc *Service) GetAllUserPosts(ctx context.Context, page int64, limit int64, userID string, isAnonymous bool) (*PaginatedAllPosts, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 100 {
		limit = 100
	}

	posts, totalPosts, err := svc.repo.FetchAllUserPosts(ctx, userID, page, limit, isAnonymous)
	if err != nil {
		return &PaginatedAllPosts{}, err
	}

	totalPages := int64(math.Ceil(float64(totalPosts) / float64(limit)))

	return &PaginatedAllPosts{
		Data:       posts,
		Page:       page,
		Limit:      limit,
		TotalPosts: totalPosts,
		TotalPages: totalPages,
	}, nil
}

func (svc *Service) GetAllPosts(ctx context.Context, page int64, limit int64) (*PaginatedAllPosts, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 100 {
		limit = 100
	}

	posts, totalPosts, err := svc.repo.FetchAllPosts(ctx, page, limit)
	if err != nil {
		return &PaginatedAllPosts{}, err
	}

	totalPages := int64(math.Ceil(float64(totalPosts) / float64(limit)))

	return &PaginatedAllPosts{
		Data:       posts,
		Page:       page,
		Limit:      limit,
		TotalPosts: totalPosts,
		TotalPages: totalPages,
	}, nil
}

func (svc *Service) GetSinglePost(ctx context.Context, postID string) (Post, error) {
	post, err := svc.repo.FindByID(ctx, postID)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (svc *Service) DeletePost(ctx context.Context, postID string, userID string, role string) (Post, error) {
	post, err := svc.repo.Delete(ctx, postID, userID, svc.config, role)
	if err != nil {
		return Post{}, fmt.Errorf("%w", err)
	}

	return post, nil
}

func (svc *Service) UpdatePost(ctx context.Context, postID string, updateForm UpdatePostForm) (Post, error) {
	title := strings.TrimSpace(updateForm.Title)
	descp := strings.TrimSpace(updateForm.Description)
	postType := strings.TrimSpace(updateForm.PostType)

	if len(title) > 100 {
		return Post{}, errors.New("title length exceeded (must be under 100 characters)")
	}

	if len(descp) > 5000 {
		return Post{}, errors.New("description length exceeded (must be under 5000 characters)")
	}

	if postType == "" {
		return Post{}, errors.New("post type not provided")
	}

	var imageURL string

	if updateForm.Image != nil {

		file, _, err := utils.ValidateImageType(updateForm.Image)
		if err != nil {
			return Post{}, err
		}

		img, _, err := image.Decode(file)
		file.Close()

		if err != nil {
			return Post{}, fmt.Errorf("invalid or corrupted image (%w)", err)
		}

		resized := utils.ResizeIfLarge(img, utils.MaxImageWidth)
		compressed, err := utils.CompressJpeg(resized, utils.MaxImageSizeBytes)
		if err != nil {
			return Post{}, fmt.Errorf("compression image failed (%w)", err)
		}

		filename := uuid.New().String() + ".jpg"
		url, err := storage.UploadImageToSupabase(compressed, "image/jpeg", string(filename), svc.config)
		if err != nil {
			return Post{}, fmt.Errorf("image upload failed (%w)", err)
		}

		imageURL = url
	}

	now := time.Now().UTC()

	postReq := UpdatePost{
		Title:       updateForm.Title,
		Description: updateForm.Description,
		PostType:    updateForm.PostType,
		ImageURL:    imageURL,
		UpdatedAt:   now,
	}

	post, err := svc.repo.Update(ctx, postID, postReq)
	if err != nil {
		return Post{}, fmt.Errorf("%w", err)
	}

	return post, nil
}
