package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserName     string             `bson:"userName,omitempty" json:"userName"`
	ProfileURL   string             `bson:"profileUrl,omitempty" json:"profileUrl"`
	Email        string             `bson:"email" json:"email"`
	Mobile       string             `bson:"mobile,omitempty" json:"mobile"`
	Name         string             `bson:"name,omitempty" json:"name"`
	Branch       string             `bson:"branch,omitempty" json:"branch"`
	Year         int                `bson:"year,omitempty" json:"year"`
	PasswordHash string             `bson:"passwordHash" json:"passwordhash"`
	Role         string             `bson:"role" json:"role"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type UpdateUser struct {
	ProfileURL string    `bson:"profileUrl,omitempty" json:"profileUrl"`
	UserName   string    `bson:"userName,omitempty" json:"userName"`
	Mobile     string    `bson:"mobile,omitempty" json:"mobile"`
	Name       string    `bson:"name,omitempty" json:"name"`
	Branch     string    `bson:"branch,omitempty" json:"branch"`
	Year       int       `bson:"year,omitempty" json:"year"`
	UpdatedAt  time.Time `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type UpdateUserForm struct {
	ProfileURL string `form:"profileUrl"`
	UserName   string `form:"userName"`
	Mobile     string `form:"mobile"`
	Name       string `form:"name"`
	Branch     string `form:"branch"`
	Year       int    `form:"year"`
}

type PaginatedUser struct {
	Data       []PublicUser `json:"users"`
	Page       int64        `json:"page"`
	Limit      int64        `json:"limit"`
	TotalUsers int64        `json:"totalUsers"`
	TotalPages int64        `json:"totalPages"`
}

type PublicUser struct {
	ID         string    `json:"id"`
	ProfileURL string    `json:"profileURL"`
	UserName   string    `json:"userName"`
	Email      string    `json:"email"`
	Mobile     string    `json:"mobile"`
	Name       string    `json:"name"`
	Branch     string    `json:"branch"`
	Year       int       `json:"year"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func ToPublic(user User) PublicUser {
	return PublicUser{
		ID:         user.ID.Hex(),
		ProfileURL: user.ProfileURL,
		UserName:   user.UserName,
		Email:      user.Email,
		Mobile:     user.Mobile,
		Name:       user.Name,
		Branch:     user.Branch,
		Year:       user.Year,
		Role:       user.Role,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}
