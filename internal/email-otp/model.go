package emailotp

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailOTP struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	OTPHash   string             `bson:"otpHash"`
	Verified  bool               `bson:"verified"`
	Attempts  int                `bson:"attempts"`
	ExpiresAt time.Time          `bson:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type SendReq struct {
	Email string `json:"email" binding:"required"`
}

type VerifyReq struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}
