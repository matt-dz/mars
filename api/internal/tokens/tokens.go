// Package tokens contains utility functions for creating api tokens
package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"mars/internal/env"
	marsjwt "mars/internal/jwt"
	"mars/internal/role"

	"github.com/google/uuid"
)

const (
	RefreshTokenBytes = 64
	CSRFTokenBytes    = 64
)

const (
	refreshTokenName = "refresh"
	accessTokenName  = "access"
	csrfTokenName    = "csrf"
)

func RefreshTokenDuration() time.Duration {
	return time.Hour * 24 * 14 // 14 days
}

func AccessTokenDuration() time.Duration {
	return time.Minute * 30 // 30 minutes
}

func CreateRefreshToken() (token string, err error) {
	bytes := make([]byte, RefreshTokenBytes)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func CreateCSRFToken() (token string, err error) {
	bytes := make([]byte, RefreshTokenBytes)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", err
	}
	// Needs to be URLEncoding otherwise weird things
	// happen on the frontend
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CreateUserAccessToken(env *env.Env, userid uuid.UUID) (token string, err error) {
	secret := env.Get("APP_SECRET")
	if secret == "" {
		return "", errors.New("no APP_SECRET set")
	}

	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decoding secret: %w", err)
	}

	jwt, err := marsjwt.GenerateJWT(marsjwt.JWTParams{
		Role:   role.RoleUser,
		UserID: userid.String(),
	}, secretBytes, "1")
	if err != nil {
		return "", fmt.Errorf("creating jwt: %w", err)
	}

	return jwt, nil
}

func NewRefreshTokenCookie(token string, secure bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     refreshTokenName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(RefreshTokenDuration().Seconds()),
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie
}

func NewAccessTokenCookie(token string, secure bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     accessTokenName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(AccessTokenDuration().Seconds()),
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie
}

func NewCSRFTokenCookie(token string, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     csrfTokenName,
		Value:    token,
		Path:     "/",
		MaxAge:   0, // will exist the duration of the session
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}
