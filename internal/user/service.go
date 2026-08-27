package user

import (
	"context"
	"errors"
	"fmt"
	"math"

	// "regexp"
	"reportit-api/internal/auth"
	"reportit-api/internal/config"
	"reportit-api/internal/storage"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo        *Repo
	refreshRepo *RefreshRepo
	config      config.Config
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Mobile   string `json:"mobile"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Mobile   string `json:"mobile"`
}

type AuthResult struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"-"`
	User         PublicUser `json:"user"`
}

func NewUserService(repo *Repo, config config.Config, refreshRepo *RefreshRepo) *Service {
	return &Service{
		repo:        repo,
		refreshRepo: refreshRepo,
		config:      config,
	}
}

// func isValidEmail(email string) bool {
// 	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

// 	re := regexp.MustCompile(pattern)

// 	return re.MatchString(email)
// }

const refreshTokenTTL = 14 * 24 * time.Hour

func (svc *Service) issueTokenPair(ctx context.Context, u User) (AuthResult, error) {
	accessToken, err := auth.CreateAccessToken(svc.config.JwtSecret, u.ID.Hex(), u.Role)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	rt := RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    u.ID,
		TokenHash: auth.HashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(refreshTokenTTL),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := svc.refreshRepo.Create(ctx, rt); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToPublic(u),
	}, nil
}

func (svc *Service) RefreshTokens(ctx context.Context, refreshToken string) (AuthResult, error) {
	if refreshToken == "" {
		return AuthResult{}, errors.New("missing refresh token")
	}

	tokenHash := auth.HashRefreshToken(refreshToken)

	stored, err := svc.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AuthResult{}, mongo.ErrNoDocuments
		}
		return AuthResult{}, err
	}

	if stored.Revoked {
		_ = svc.refreshRepo.RevokeAllForUser(ctx, stored.UserID)
		return AuthResult{}, errors.New("refresh token revoked, please log in again")
	}

	if time.Now().After(stored.ExpiresAt) {
		return AuthResult{}, errors.New("refresh token expired, please log in again")
	}

	if err := svc.refreshRepo.Revoke(ctx, tokenHash); err != nil {
		return AuthResult{}, err
	}

	u, err := svc.repo.FetchSingleUser(ctx, stored.ID.Hex())
	if err != nil {
		return AuthResult{}, err
	}

	return svc.issueTokenPair(ctx, u)
}

func (svc *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	return svc.refreshRepo.Revoke(ctx, auth.HashRefreshToken(refreshToken))
}

func isValidCollegeEmail(email string) bool {
	return strings.HasSuffix(email, "@glbitm.ac.in")
}

func containsSpecialCharacter(password string) bool {
	for _, char := range password {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// func isValidPhone(phone string) bool {
// 	phoneRegex := regexp.MustCompile(`^(?:\+91)?[6-9][0-9]{9}$`)
// 	return phoneRegex.MatchString(phone)
// }

func (svc *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.ToLower(strings.TrimSpace(input.Password))
	// mobile := strings.TrimSpace(input.Mobile)

	// error handling
	if email == "" || password == "" {
		return AuthResult{}, errors.New("email and password are required")
	}

	_, err := svc.repo.FindByEmail(ctx, email)
	if err == nil {
		return AuthResult{}, errors.New("email is already registered with other account, try a different email")
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return AuthResult{}, err
	}

	if !isValidCollegeEmail(email) {
		return AuthResult{}, errors.New("invalid email, must be a college email (domain: @glbitm.ac.in)")
	}

	if len(password) < 8 || len(password) > 16 {
		return AuthResult{}, errors.New("password length must be between 8 and 16")
	}

	if !containsSpecialCharacter(password) {
		return AuthResult{}, errors.New("password must contains a special character")
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hashing of the password failed (%w)", err)
	}

	// if !isValidPhone(mobile) {
	// 	return AuthResult{}, errors.New("invalid mobile number")
	// }

	now := time.Now()

	user := User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		PasswordHash: string(hashBytes),
		// Mobile:       mobile,
		Role:      "user",
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := svc.repo.Create(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	result, err := svc.issueTokenPair(ctx, created)
	if err != nil {
		return AuthResult{}, err
	}

	return result, nil
}

func (svc *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := strings.ToLower(strings.TrimSpace(input.Password))

	if email == "" || password == "" {
		return AuthResult{}, errors.New("email and password are required")
	}

	user, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AuthResult{}, errors.New("Invalid credentials")
		}
		return AuthResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, errors.New("invalid credentials or wrong password")
	}

	result, err := svc.issueTokenPair(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	return result, nil
}

func (svc *Service) UpdateUser(ctx context.Context, user UpdateUserForm, userID string) (PublicUser, error) {
	now := time.Now().UTC()

	update := UpdateUser{
		ProfileURL: user.ProfileURL,
		UserName:   user.UserName,
		Mobile:     user.Mobile,
		Name:       user.Name,
		Branch:     user.Branch,
		Year:       user.Year,
		UpdatedAt:  now,
	}

	updatedUser, err := svc.repo.Update(ctx, update, userID)
	if err != nil {
		return PublicUser{}, fmt.Errorf("%w", err)
	}

	return ToPublic(updatedUser), nil
}

func (svc *Service) GetAllUsers(ctx context.Context, page int64, limit int64) (*PaginatedUser, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 500 {
		limit = 500
	}

	users, totalUsers, err := svc.repo.FetchAllUser(ctx, page, limit)
	if err != nil {
		return &PaginatedUser{}, err
	}

	publicUser := make([]PublicUser, 0, len(users))

	for _, user := range users {
		publicUser = append(publicUser, ToPublic(user))
	}

	totalPages := int64(math.Ceil(float64(totalUsers) / float64(limit)))

	return &PaginatedUser{
		Data:       publicUser,
		Page:       page,
		Limit:      limit,
		TotalUsers: totalUsers,
		TotalPages: totalPages,
	}, nil
}

func (svc *Service) GetSingleUser(ctx context.Context, userID string) (PublicUser, error) {
	userDetails, err := svc.repo.FetchSingleUser(ctx, userID)

	if err != nil {
		return PublicUser{}, fmt.Errorf("%w", err)
	}

	return ToPublic(userDetails), nil
}

func (svc *Service) DeleteUser(ctx context.Context, userID string) (PublicUser, error) {
	userDetails, err := svc.repo.Delete(ctx, userID)
	if err != nil {
		return PublicUser{}, fmt.Errorf("%w", err)
	}

	profileUrl := userDetails.ProfileURL
	filename := storage.FilenameFromURL(profileUrl, svc.config.SupabaseBucket)
	if err := storage.DeleteImageFromSupabase(filename, svc.config); err != nil {
		return PublicUser{}, fmt.Errorf("unable to delete the profile url (%w)", err)
	}

	return ToPublic(userDetails), nil
}
