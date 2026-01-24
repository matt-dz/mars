package openapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/argon2id"
	"mars/internal/database"
	"mars/internal/tokens"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type loginSuccessResponse struct {
	accessCookie  *http.Cookie
	refreshCookie *http.Cookie
	csrfCookie    *http.Cookie
	body          LoginResponse
}

func (r loginSuccessResponse) VisitPostApiLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, r.accessCookie)
	http.SetCookie(w, r.refreshCookie)
	http.SetCookie(w, r.csrfCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	return encoder.Encode(r.body)
}

func (s Server) PostApiLogin(ctx context.Context, request PostApiLoginRequestObject) (
	PostApiLoginResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	// Get user from database
	s.Env.Logger.DebugContext(ctx, "getting user")
	user, err := s.Env.Database.GetUserByEmail(ctx, string(request.Body.Email))
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "user with email does not exist", slog.Any("error", err))
		return PostApiLogin401JSONResponse{
			Message: "invalid email or password",
			ErrorId: reqid,
			Code:    apierror.InvalidCredentials.String(),
			Status:  apierror.InvalidCredentials.Status(),
		}, nil
	}
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get user", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	// Decode ground password hash
	s.Env.Logger.DebugContext(ctx, "decoding password hash")
	hashParams, hashSalt, groundHash, err := argon2id.DecodeHash(user.PasswordHash)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode password hash", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	// Hash given password
	s.Env.Logger.DebugContext(ctx, "hashing given password")
	givenHash := argon2id.HashWithSalt(request.Body.Password, *hashParams, hashSalt)

	// Compare passwords
	s.Env.Logger.DebugContext(ctx, "comparing passwords")
	if subtle.ConstantTimeCompare(givenHash, groundHash) == 0 {
		s.Env.Logger.ErrorContext(ctx, "passwords do not match")
		return PostApiLogin401JSONResponse{
			Message: "invalid email or password",
			ErrorId: reqid,
			Code:    apierror.InvalidCredentials.String(),
			Status:  apierror.InvalidCredentials.Status(),
		}, nil
	}

	s.Env.Logger.DebugContext(ctx, "creating tokens")

	// Create refresh token
	refresh, err := tokens.CreateRefreshToken()
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create refresh token", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}
	refreshHash, err := argon2id.HashAndEncode(refresh, argon2id.DefaultParams)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to hash refresh token", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}
	s.Env.Database.UpdateUserRefreshToken(ctx, database.UpdateUserRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{
			String: refreshHash,
			Valid:  true,
		},
		RefreshTokenExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(tokens.RefreshTokenDuration()),
			Valid: true,
		},
	})

	// Create CSRF token
	csrf, err := tokens.CreateCSRFToken()
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create csrf token", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	// Create access token
	access, err := tokens.CreateUserAccessToken(s.Env, user.ID)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create user access token", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	// Return response
	return loginSuccessResponse{
		accessCookie:  tokens.NewAccessTokenCookie(access, s.Env.IsProd()),
		refreshCookie: tokens.NewRefreshTokenCookie(refresh, s.Env.IsProd()),
		csrfCookie:    tokens.NewCSRFTokenCookie(csrf, s.Env.IsProd()),
		body: LoginResponse{
			AccessToken: access,
			ExpiresIn:   int64(tokens.AccessTokenDuration().Seconds()),
			TokenType:   "bearer",
		},
	}, nil
}
