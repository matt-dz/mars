package openapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"
	"mars/internal/tokens"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) PostApiOauthSpotifyToken(ctx context.Context, request PostApiOauthSpotifyTokenRequestObject) (
	PostApiOauthSpotifyTokenResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userid, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Exchange code for token
	s.Env.Logger.DebugContext(ctx, "exchanging code for tokens")
	endpoint := "https://accounts.spotify.com/api/token"
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("code", request.Body.Code)
	form.Add("redirect_uri", os.Getenv("SPOTIFY_REDIRECT_URI"))
	body := form.Encode()
	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create request", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	clientid := os.Getenv("SPOTIFY_CLIENT_ID")
	clientsecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientid == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_CLIENT_ID not set")
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if clientsecret == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_CLIENT_SECRET not set")
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientid+":"+clientsecret))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authorization)

	resp, err := s.Env.HTTP.Do(req)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to send exchange request", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.Env.Logger.ErrorContext(ctx, "exchange request failed with non-200 status", slog.String("body", string(body)))
		return PostApiOauthSpotifyToken400JSONResponse{
			Message: "bad request",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	type oauthTokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
	}
	var oauthTokens oauthTokenResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&oauthTokens)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode exchange response body", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get current user profile
	endpoint = "https://api.spotify.com/v1/me"
	s.Env.Logger.DebugContext(ctx, "getting user spotify profile")
	type profileResponse struct {
		ID string `json:"id"`
		// more fields, but idc about them
	}
	req, err = retryablehttp.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create profile request", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	req.Header.Set("Authorization", "Bearer "+oauthTokens.AccessToken)
	resp, err = s.Env.HTTP.Do(req)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to send profile request", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.Env.Logger.ErrorContext(ctx, "profile request failed with non-200 status", slog.String("body", string(body)))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	var profile profileResponse
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&profile)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode profile response body", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Update user spotify id
	s.Env.Logger.DebugContext(ctx, "updating user spotify id")
	err = s.Env.Database.UpdateUserSpotifyID(ctx, database.UpdateUserSpotifyIDParams{
		SpotifyID: pgtype.Text{
			String: profile.ID,
			Valid:  true,
		},
		ID: userid,
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to update user spotify id", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Update tokens
	s.Env.Logger.DebugContext(ctx, "updating user spotify tokens")
	expiration := time.Duration(oauthTokens.ExpiresIn) * time.Second
	err = s.Env.Database.UpsertUserSpotifyTokens(ctx, database.UpsertUserSpotifyTokensParams{
		SpotifyUserID: profile.ID,
		AccessToken:   oauthTokens.AccessToken,
		TokenType:     oauthTokens.TokenType,
		RefreshToken:  oauthTokens.RefreshToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(expiration),
			Valid: true,
		},
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to update user spotify tokens", slog.Any("error", err))
		return PostApiOauthSpotifyToken500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return PostApiOauthSpotifyToken204Response{}, nil
}

func (s Server) GetApiSpotifyStatus(
	ctx context.Context, request GetApiSpotifyStatusRequestObject) (
	GetApiSpotifyStatusResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userid, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return GetApiSpotifyStatus500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get connection status
	s.Env.Logger.DebugContext(ctx, "getting connection status")
	expiration, err := s.Env.Database.GetUserSpotifyTokenExpiration(ctx, userid)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "no rows found", slog.Any("error", err))
		return GetApiSpotifyStatus200JSONResponse{
			Connected: false,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get status", slog.Any("error", err))
		return GetApiSpotifyStatus500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if !expiration.Valid || expiration.Time.Before(time.Now()) {
		s.Env.Logger.ErrorContext(ctx, "tokens have expired")
		return GetApiSpotifyStatus200JSONResponse{
			Connected: false,
		}, nil
	}

	return GetApiSpotifyStatus200JSONResponse{
		Connected: true,
	}, nil
}

func (s Server) PostApiOauthSpotifyTokenRefresh(
	ctx context.Context, request PostApiOauthSpotifyTokenRefreshRequestObject) (
	PostApiOauthSpotifyTokenRefreshResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	// Get user Spotify refresh token
	s.Env.Logger.DebugContext(ctx, "getting spotify refresh tokens")
	refreshToken, err := s.Env.Database.GetUserSpotifyRefreshToken(ctx, request.Body.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "no rows returned, user has no spotify integration", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh404JSONResponse{
			Message: "no spotify integration found",
			Status:  apierror.NoSpotifyIntegration.Status(),
			Code:    apierror.NoSpotifyIntegration.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get spotify refresh tokens", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Refresh tokens
	s.Env.Logger.DebugContext(ctx, "refreshing spotify tokens")
	const endpoint = "https://accounts.spotify.com/api/token"
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("refresh_token", refreshToken)
	body := form.Encode()
	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create request", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	clientid := os.Getenv("SPOTIFY_CLIENT_ID")
	clientsecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientid == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_CLIENT_ID not set")
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if clientsecret == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_CLIENT_SECRET not set")
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientid+":"+clientsecret))
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authorization)

	res, err := s.Env.HTTP.Do(req)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to send exchange request", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		s.Env.Logger.ErrorContext(ctx, "refresh request failed with non-200 status", slog.String("body", string(body)))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	var refreshResponse struct {
		ExpiresIn    int    `json:"expires_in"`
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	err = json.NewDecoder(res.Body).Decode(&refreshResponse)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode spotify response", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if refreshResponse.RefreshToken == "" {
		refreshResponse.RefreshToken = refreshToken
	}

	// Update tokens
	s.Env.Logger.DebugContext(ctx, "updating spotify tokens in database")
	userid, err := s.Env.Database.GetUserSpotifyId(ctx, request.Body.UserId)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get user spotify id", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	expiration := time.Duration(refreshResponse.ExpiresIn) * time.Second
	err = s.Env.Database.UpdateUserSpotifyTokens(ctx, database.UpdateUserSpotifyTokensParams{
		AccessToken:  refreshResponse.AccessToken,
		RefreshToken: refreshResponse.RefreshToken,
		TokenType:    refreshResponse.TokenType,
		Scope:        refreshResponse.Scope,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(expiration),
			Valid: true,
		},
		SpotifyUserID: userid.String,
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to update tokens in database", slog.Any("error", err))
		return PostApiOauthSpotifyTokenRefresh500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return PostApiOauthSpotifyTokenRefresh204Response{}, nil
}

func (s Server) GetApiOauthSpotifyConfigJson(
	ctx context.Context, request GetApiOauthSpotifyConfigJsonRequestObject) (
	GetApiOauthSpotifyConfigJsonResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	clientid := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")

	if clientid == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_CLIENT_ID not set")
		return GetApiOauthSpotifyConfigJson500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	if redirectURI == "" {
		s.Env.Logger.ErrorContext(ctx, "SPOTIFY_REDIRECT_URI not set")
		return GetApiOauthSpotifyConfigJson500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return GetApiOauthSpotifyConfigJson200JSONResponse{
		ResponseType: "code",
		ClientId:     clientid,
		RedirectUri:  redirectURI,
		Scope: "user-read-private user-read-email user-library-read " +
			"user-top-read user-read-recently-played playlist-modify-public " +
			"playlist-modify-private ugc-image-upload",
	}, nil
}
