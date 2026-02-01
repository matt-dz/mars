package openapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"
	"mars/internal/log"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) PostApiIntegrationsSpotifyTracksSync(
	ctx context.Context, request PostApiIntegrationsSpotifyTracksSyncRequestObject) (
	PostApiIntegrationsSpotifyTracksSyncResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	ctx = log.AppendCtx(ctx, slog.String("user-id", request.Body.UserId.String()))

	// Get user spotify tokens
	s.Env.Logger.DebugContext(ctx, "getting spotify access token")
	accessToken, err := s.Env.Database.GetUserSpotifyAccessToken(ctx, request.Body.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "no rows returned - no spotify integration", slog.Any("error", err))
		return PostApiIntegrationsSpotifyTracksSync404JSONResponse{
			Message: "user does not have a spotify integration",
			Status:  apierror.NoSpotifyIntegration.Status(),
			Code:    apierror.NoSpotifyIntegration.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get access token", slog.Any("error", err))
		return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get recent tracks
	s.Env.Logger.DebugContext(ctx, "getting recently played tracks")
	const endpoint = "https://api.spotify.com/v1/me/player/recently-played"
	req, err := retryablehttp.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create request", slog.Any("error", err))
		return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := s.Env.HTTP.Do(req)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to send request", slog.Any("error", err))
		return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		s.Env.Logger.ErrorContext(ctx, "request failed with non-200 status", slog.String("body", string(body)))
		return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	var body struct {
		Items []struct {
			Track struct {
				Album struct {
					Images []struct {
						Url string `json:"url"`
					} `json:"images"`
				} `json:"album"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				ID           string `json:"id"`
				Name         string `json:"name"`
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"track"`
			PlayedAt time.Time `json:"played_at"`
		} `json:"items"`
	}
	if err = json.NewDecoder(res.Body).Decode(&body); err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to decode response body", slog.Any("error", err))
		return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Upload recent tracks
	s.Env.Logger.DebugContext(ctx, "uploading tracks")
	for _, item := range body.Items {

		// Upsert track
		artists := make([]string, len(item.Track.Artists))
		for i, artist := range item.Track.Artists {
			artists[i] = artist.Name
		}
		var imageURL string
		if len(item.Track.Album.Images) > 0 {
			imageURL = item.Track.Album.Images[0].Url
		}
		err = s.Env.Database.UpsertTrack(ctx, database.UpsertTrackParams{
			ID:      item.Track.ID,
			Name:    item.Track.Name,
			Href:    item.Track.ExternalUrls.Spotify,
			Artists: artists,
			ImageUrl: pgtype.Text{
				String: imageURL,
				Valid:  imageURL != "",
			},
		})
		if err != nil {
			s.Env.Logger.ErrorContext(ctx, "failed to upsert track",
				slog.Group("track", slog.String("id", item.Track.ID)), slog.Any("error", err))
			return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
				Message: "internal server error",
				Status:  apierror.InternalServerError.Status(),
				Code:    apierror.InternalServerError.String(),
				ErrorId: reqid,
			}, nil
		}

		// Create listen
		err = s.Env.Database.UpsertTrackListen(ctx, database.UpsertTrackListenParams{
			UserID:  request.Body.UserId,
			TrackID: item.Track.ID,
			PlayedAt: pgtype.Timestamptz{
				Time:  item.PlayedAt,
				Valid: true,
			},
		})
		if err != nil {
			s.Env.Logger.ErrorContext(ctx, "failed to create listen",
				slog.Group("track", slog.String("id", item.Track.ID)), slog.Any("error", err))
			return PostApiIntegrationsSpotifyTracksSync500JSONResponse{
				Message: "internal server error",
				Status:  apierror.InternalServerError.Status(),
				Code:    apierror.InternalServerError.String(),
				ErrorId: reqid,
			}, nil
		}
	}
	return PostApiIntegrationsSpotifyTracksSync204Response{}, nil
}
