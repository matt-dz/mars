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

type refreshSessionSuccessResponse struct {
	accessCookie  *http.Cookie
	refreshCookie *http.Cookie
	csrfCookie    *http.Cookie
	body          LoginResponse
}

func (r refreshSessionSuccessResponse) VisitPostApiAuthRefreshResponse(w http.ResponseWriter) error {
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
	refresh, err := tokens.CreateRefreshToken(user.ID)
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
	err = s.Env.Database.UpdateUserRefreshToken(ctx, database.UpdateUserRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{
			String: refreshHash,
			Valid:  true,
		},
		RefreshTokenExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(tokens.RefreshTokenDuration()),
			Valid: true,
		},
		ID: user.ID,
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to update user refresh token", slog.Any("error", err))
		return PostApiLogin500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

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
			TokenType:   "Bearer",
		},
	}, nil
}

func (s Server) PostApiAuthRefresh(ctx context.Context, request PostApiAuthRefreshRequestObject) (
	PostApiAuthRefreshResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	// Get refresh token
	s.Env.Logger.DebugContext(ctx, "getting refresh token")
	var refreshToken string
	if request.Body != nil && request.Body.RefreshToken != nil {
		refreshToken = *request.Body.RefreshToken
	} else if request.Params.Refresh != nil {
		refreshToken = *request.Params.Refresh
	}
	if refreshToken == "" {
		s.Env.Logger.ErrorContext(ctx, "refresh token not provided")
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "refresh token not provided",
			ErrorId: reqid,
		}, nil
	}

	// Parse refresh token
	s.Env.Logger.DebugContext(ctx, "parsing refresh token")
	userid, err := tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to parse refresh token", slog.Any("error", err))
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "invalid refresh token",
			ErrorId: reqid,
		}, nil
	}

	// Get current refresh token
	s.Env.Logger.DebugContext(ctx, "getting user refresh token from db")
	refresh, err := s.Env.Database.GetUserRefreshToken(ctx, userid)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "failed to find user", slog.Any("error", err))
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "invalid refresh token",
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		return PostApiAuthRefresh500JSONResponse{
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			Message: "internal server error",
			ErrorId: reqid,
		}, nil
	}
	if !refresh.RefreshTokenHash.Valid {
		s.Env.Logger.ErrorContext(ctx, "user does not have a refresh token", slog.Any("error", err))
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "invalid refresh token",
			ErrorId: reqid,
		}, nil
	}
	if refresh.RefreshTokenExpiresAt.Time.Before(time.Now()) {
		s.Env.Logger.ErrorContext(ctx, "refresh token has expired")
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "invalid refresh token",
			ErrorId: reqid,
		}, nil
	}

	// Compare tokens
	s.Env.Logger.DebugContext(ctx, "comparing refresh tokens")
	argonParams, salt, groundRefreshHash, err := argon2id.DecodeHash(refresh.RefreshTokenHash.String)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode refresh token hash", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			Message: "internal server error",
			ErrorId: reqid,
		}, nil
	}
	givenRefreshHash := argon2id.HashWithSalt(refreshToken, *argonParams, salt)
	if subtle.ConstantTimeCompare(givenRefreshHash, groundRefreshHash) == 0 {
		s.Env.Logger.ErrorContext(ctx, "refresh tokens do not match", slog.Any("error", err))
		return PostApiAuthRefresh401JSONResponse{
			Status:  apierror.InvalidRefreshToken.Status(),
			Code:    apierror.InvalidRefreshToken.String(),
			Message: "invalid refresh token",
			ErrorId: reqid,
		}, nil
	}

	// Create new token
	s.Env.Logger.DebugContext(ctx, "creating new tokens")
	newToken, err := tokens.CreateRefreshToken(userid)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create new refresh token", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			Message: "internal server error",
			ErrorId: reqid,
		}, nil
	}
	newTokenHash, err := argon2id.HashAndEncode(newToken, argon2id.DefaultParams)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to hash refresh token", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			Message: "internal server error",
			ErrorId: reqid,
		}, nil
	}
	err = s.Env.Database.UpdateUserRefreshToken(ctx, database.UpdateUserRefreshTokenParams{
		RefreshTokenHash: pgtype.Text{
			String: newTokenHash,
			Valid:  true,
		},
		RefreshTokenExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(tokens.AccessTokenDuration()),
			Valid: true,
		},
		ID: userid,
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to update refresh token", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			Message: "internal server error",
			ErrorId: reqid,
		}, nil
	}

	csrf, err := tokens.CreateCSRFToken()
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create csrf token", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	access, err := tokens.CreateUserAccessToken(s.Env, userid)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create user access token", slog.Any("error", err))
		return PostApiAuthRefresh500JSONResponse{
			Message: "Internal Server Error",
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	// Return response
	return refreshSessionSuccessResponse{
		accessCookie:  tokens.NewAccessTokenCookie(access, s.Env.IsProd()),
		refreshCookie: tokens.NewRefreshTokenCookie(newToken, s.Env.IsProd()),
		csrfCookie:    tokens.NewCSRFTokenCookie(csrf, s.Env.IsProd()),
		body: LoginResponse{
			AccessToken: access,
			ExpiresIn:   int64(tokens.AccessTokenDuration().Seconds()),
			TokenType:   "Bearer",
		},
	}, nil
}

func (s Server) GetApiAuthVerify(
	ctx context.Context, request GetApiAuthVerifyRequestObject,
) (GetApiAuthVerifyResponseObject, error) {
	return GetApiAuthVerify204Response{}, nil
}
