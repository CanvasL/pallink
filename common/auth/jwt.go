package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateToken creates a JWT for the given user id.
func GenerateToken(secret string, expireSeconds int64, userID uint64) (string, error) {
	if expireSeconds <= 0 {
		expireSeconds = 86400
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"iat":    now.Unix(),
		"exp":    now.Add(time.Duration(expireSeconds) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
