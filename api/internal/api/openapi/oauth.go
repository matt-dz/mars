package openapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
