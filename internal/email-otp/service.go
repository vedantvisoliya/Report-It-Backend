package emailotp

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"reportit-api/internal/services"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo             *Repo
	gmailAppPassword string
}

func NewOTPService(repo *Repo, gmailAppPassword string) *Service {
	return &Service{
		repo:             repo,
		gmailAppPassword: gmailAppPassword,
	}
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func (svc *Service) SendOTP(ctx context.Context, email string) error {
	otp, err := generateOTP()
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := svc.repo.DeleteByEmail(ctx, email); err != nil {
		return err
	}

	record := EmailOTP{
		ID:        primitive.NewObjectID(),
		Email:     email,
		OTPHash:   string(hash),
		Verified:  false,
		Attempts:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := svc.repo.Create(ctx, record); err != nil {
		return err
	}

	err = services.SendOTPEmail(email, otp, svc.gmailAppPassword)
	if err != nil {
		return fmt.Errorf("failed to send OTP (%w)", err)
	}

	return nil
}

func (svc *Service) VerifyOTP(ctx context.Context, email string, otp string) error {
	record, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("OTP expired or not found")
	}

	if record.Attempts >= 5 {
		svc.repo.DeleteByID(ctx, record.ID.Hex())
		return errors.New("too many attempts")
	}

	if time.Now().After(record.ExpiresAt) {
		svc.repo.DeleteByID(ctx, record.ID.Hex())
		return errors.New("OTP expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.OTPHash), []byte(otp)); err != nil {
		return errors.New("Invalid OTP")
	}

	svc.repo.DeleteByID(ctx, record.ID.Hex())
	return nil
}
