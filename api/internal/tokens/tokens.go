// Package tokens contains utility functions for creating api tokens
package tokens

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mars/internal/env"
	marsjwt "mars/internal/jwt"
	"mars/internal/role"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	RefreshTokenBytes = 64
	CSRFTokenBytes    = 64
)

const (
	RefreshTokenName    = "refresh"
	AccessTokenName     = "access"
	CsrfTokenName       = "csrf"
	AuthorizationHeader = "Authorization"
	CsrfTokenHeader     = "X-CSRF-Token"
)

type (
	useridCtxKeyType      struct{}
	accessTokenCtxKeyType struct{}
)

var (
	useridCtxKey      useridCtxKeyType
	accessTokenCtxKey accessTokenCtxKeyType
)

func RefreshTokenDuration() time.Duration {
	return time.Hour * 24 * 14 // 14 days
}

func AccessTokenDuration() time.Duration {
	return time.Minute * 30 // 30 minutes
}

func CreateRefreshToken(userid uuid.UUID) (token string, err error) {
	bytes := make([]byte, RefreshTokenBytes)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s$%s", userid, base64.URLEncoding.EncodeToString(bytes)), nil
}

func ParseRefreshToken(refreshtoken string) (
	userid uuid.UUID, err error,
) {
	id, _, found := strings.Cut(refreshtoken, "$")
	if !found {
		return userid, errors.New("invalid refresh token, expected format \"<user-id>$<random>\"")
	}
	userid, err = uuid.Parse(id)
	if err != nil {
		return userid, fmt.Errorf("invalid user id: %w", err)
	}
	return userid, nil
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

func CreateAccessToken(env *env.Env, userid uuid.UUID, role role.Role) (token string, err error) {
	secret := env.Get("APP_SECRET")
	if secret == "" {
		return "", errors.New("APP_SECRET not set")
	}

	jwt, err := marsjwt.GenerateJWT(marsjwt.JWTParams{
		Role:   role,
		UserID: userid.String(),
	}, []byte(secret), "1", AccessTokenDuration())
	if err != nil {
		return "", fmt.Errorf("creating jwt: %w", err)
	}

	return jwt, nil
}

func NewRefreshTokenCookie(token string, secure bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     RefreshTokenName,
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
		Name:     AccessTokenName,
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
		Name:     CsrfTokenName,
		Value:    token,
		Path:     "/",
		MaxAge:   0, // will exist the duration of the session
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func ParseBearerToken(bearertoken string) (string, error) {
	token, found := strings.CutPrefix(bearertoken, "Bearer ")
	if !found {
		return "", errors.New("bearer token should be in format \"Bearer <token>\"")
	}
	return token, nil
}

func UserIDWithContext(ctx context.Context, userid uuid.UUID) context.Context {
	return context.WithValue(ctx, useridCtxKey, userid)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userid, ok := ctx.Value(useridCtxKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("invalid type")
	}
	return userid, nil
}

func AccessTokenWithContext(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, accessTokenCtxKey, token)
}

func AccessTokenFromContext(ctx context.Context) (*jwt.Token, error) {
	accessToken, ok := ctx.Value(useridCtxKey).(jwt.Token)
	if !ok {
		return nil, errors.New("invalid type")
	}
	return &accessToken, nil
}
