package auth

import (
	"errors"
	"strings"
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

// ParseUserIDFromToken validates a JWT and extracts the userId claim.
func ParseUserIDFromToken(secret, tokenString string) (uint64, error) {
	tokenString = strings.TrimSpace(tokenString)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimPrefix(tokenString, "bearer ")
	if tokenString == "" {
		return 0, errors.New("token required")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	userID, ok := parseUserID(claims["userId"])
	if !ok || userID == 0 {
		return 0, errors.New("invalid userId claim")
	}
	return userID, nil
}
