// Package jwt provides functions for generating and validating JWTs
package jwt

import (
	"fmt"
	"time"

	"mars/internal/role"

	"github.com/golang-jwt/jwt/v5"
)

type JWTParams struct {
	Role   role.Role
	UserID string
}

const (
	DefaultKID = "1"
)

func GenerateJWT(params JWTParams, secret []byte, version string, duration time.Duration) (string, error) {
	// Build token
	claims := jwt.MapClaims{
		"sub":  params.UserID,
		"role": params.Role.String(),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = version

	// Sign token
	signedKey, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signedKey, nil
}

func ValidateJWT(rawToken, version string, secret []byte) (*jwt.Token, error) {
	parserFunc := func(token *jwt.Token) (any, error) {
		kidVal, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing/invalid kid value")
		}

		if kidVal != version {
			return nil, fmt.Errorf("verifying KID value, value=%q", kidVal)
		}

		return secret, nil
	}

	// Parse the token
	token, err := jwt.Parse(rawToken, parserFunc)
	if err != nil {
		return nil, err
	}

	return token, nil
}
