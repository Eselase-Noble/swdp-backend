package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate jwt token for a particular user
func Generate(userId, role, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userId,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 8).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
