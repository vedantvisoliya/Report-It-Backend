package post

import (
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostedByID  primitive.ObjectID `bson:"postedBy" json:"postedBy"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	ImageURL    string             `bson:"imageUrl,omitempty" json:"imageUrl"`
	PostType    string             `bson:"postType" json:"postType"`
	EditHistory []UpdatePost       `bson:"editHistory,omitempty" json:"editHistory"`
	IsAnonymous bool               `bson:"isAnonymous" json:"isAnonymous"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type UpdatePostForm struct {
	Title       string                `form:"title"`
	Description string                `form:"description"`
	PostType    string                `form:"postType"`
	Image       *multipart.FileHeader `form:"image"`
}

type UpdatePost struct {
	Title       string    `bson:"title,omitempty" json:"title"`
	Description string    `bson:"description,omitempty" json:"description"`
	PostType    string    `bson:"postType,omitempty" json:"postType"`
	ImageURL    string    `bson:"imageUrl,omitempty" json:"imageUrl"`
	UpdatedAt   time.Time `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type PaginatedUserPosts struct {
	Data       []Post `json:"posts"`
	Page       int64  `json:"page"`
	Limit      int64  `json:"limit"`
	TotalPosts int64  `json:"totalPosts"`
	TotalPages int64  `json:"totalPages"`
}

type PaginatedAllPosts struct {
	Data       []Post `json:"posts"`
	Page       int64  `json:"page"`
	Limit      int64  `json:"limit"`
	TotalPosts int64  `json:"totalPosts"`
	TotalPages int64  `json:"totalPages"`
}
