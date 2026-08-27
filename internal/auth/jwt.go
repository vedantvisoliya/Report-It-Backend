package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func CreateToken(jwtSecret string, userID string, role string) (string, error) {
	now := time.Now()
	expireTime := now.Add(14 * 24 * time.Hour)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireTime),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign in token failed (%w)", err)
	}

	return signed, nil
}

func ParseToken(jwtSecret string, tokenString string) (Claims, error) {
	var claims Claims

	parsed, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method (%v)", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return Claims{}, fmt.Errorf("Parse token failed (%w)", err)
	}

	if !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}

	if claims.Subject == "" {
		return Claims{}, errors.New("token missing subject")
	}

	return claims, nil
}
