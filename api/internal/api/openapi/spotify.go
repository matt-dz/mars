package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"
	"mars/internal/log"
	"mars/internal/tokens"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// spotifyPlaylistResult represents a created Spotify playlist.
type spotifyPlaylistResult struct {
	ID  string
	URL string
}

// createSpotifyPlaylist creates a new playlist on Spotify for the given user.
func (s Server) createSpotifyPlaylist(
	_ context.Context,
	accessToken string,
	spotifyUserID string,
	name string,
) (*spotifyPlaylistResult, error) {
	endpoint := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", spotifyUserID)
	body, err := json.Marshal(map[string]any{
		"name":          name,
		"public":        true,
		"collaborative": false,
		"description":   "",
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := s.Env.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("spotify returned status %d: %s", res.StatusCode, string(respBody))
	}

	var playlistResp struct {
		ID           string `json:"id"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
	}
	if err = json.NewDecoder(res.Body).Decode(&playlistResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &spotifyPlaylistResult{
		ID:  playlistResp.ID,
		URL: playlistResp.ExternalUrls.Spotify,
	}, nil
}

// addTracksToSpotifyPlaylist adds tracks to an existing Spotify playlist.
func (s Server) addTracksToSpotifyPlaylist(
	_ context.Context,
	accessToken string,
	spotifyPlaylistID string,
	trackURIs []string,
) error {
	endpoint := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", spotifyPlaylistID)
	body, err := json.Marshal(map[string]any{
		"uris":     trackURIs,
		"position": 0,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := s.Env.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("spotify returned status %d: %s", res.StatusCode, string(respBody))
	}

	return nil
}

// getSpotifyCredentials retrieves the Spotify user ID and access token for a user.
func (s Server) getSpotifyCredentials(
	ctx context.Context, userID uuid.UUID) (
	spotifyUserID string, accessToken string, err error,
) {
	spotifyID, err := s.Env.Database.GetUserSpotifyId(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("get spotify user id: %w", err)
	}

	token, err := s.Env.Database.GetUserSpotifyAccessToken(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("get access token: %w", err)
	}

	return spotifyID.String, token, nil
}

// getPlaylistTrackURIs retrieves track URIs for a playlist.
func (s Server) getPlaylistTrackURIs(ctx context.Context, playlistID uuid.UUID) ([]string, error) {
	tracks, err := s.Env.Database.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	uris := make([]string, len(tracks))
	for i, t := range tracks {
		uris[i] = t.Uri
	}
	return uris, nil
}

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
	const endpoint = "https://api.spotify.com/v1/me/player/recently-played?limit=50"
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
						URL string `json:"url"`
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
				URI string `json:"uri"`
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
			imageURL = item.Track.Album.Images[0].URL
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
			Uri: item.Track.URI,
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

func (s Server) PostApiIntegrationsSpotifyPlaylist(
	ctx context.Context, request PostApiIntegrationsSpotifyPlaylistRequestObject) (
	PostApiIntegrationsSpotifyPlaylistResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	ctx = log.AppendCtx(ctx, slog.String("user-id", request.Body.UserId.String()))
	ctx = log.AppendCtx(ctx, slog.String("playlist-id", request.Body.PlaylistId.String()))

	// Get Spotify credentials
	s.Env.Logger.DebugContext(ctx, "getting spotify credentials")
	spotifyUserID, accessToken, err := s.getSpotifyCredentials(ctx, request.Body.UserId)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "user does not have a spotify integration", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist404JSONResponse{
			Message: "user does not have a spotify integration",
			Status:  apierror.NoSpotifyIntegration.Status(),
			Code:    apierror.NoSpotifyIntegration.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get spotify credentials", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get playlist from database
	playlist, err := s.Env.Database.GetUserPlaylist(ctx, database.GetUserPlaylistParams{
		UserID: request.Body.UserId,
		ID:     request.Body.PlaylistId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "playlist not found", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist404JSONResponse{
			Message: "playlist not found",
			Status:  apierror.PlaylistNotFound.Status(),
			Code:    apierror.PlaylistNotFound.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get track URIs
	trackURIs, err := s.getPlaylistTrackURIs(ctx, request.Body.PlaylistId)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlist tracks", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Create Spotify playlist
	s.Env.Logger.DebugContext(ctx, "creating spotify playlist")
	spotifyPlaylist, err := s.createSpotifyPlaylist(ctx, accessToken, spotifyUserID, playlist.Name)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create spotify playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	ctx = log.AppendCtx(ctx, slog.String("spotify-playlist-id", spotifyPlaylist.ID))

	// Add tracks to playlist
	s.Env.Logger.DebugContext(ctx, "adding tracks to spotify playlist")
	if err = s.addTracksToSpotifyPlaylist(ctx, accessToken, spotifyPlaylist.ID, trackURIs); err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to add tracks to spotify playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylist500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return PostApiIntegrationsSpotifyPlaylist201JSONResponse{
		Id:  spotifyPlaylist.ID,
		Url: spotifyPlaylist.URL,
	}, nil
}

func (s Server) PostApiIntegrationsSpotifyPlaylistId(
	ctx context.Context, request PostApiIntegrationsSpotifyPlaylistIdRequestObject) (
	PostApiIntegrationsSpotifyPlaylistIdResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userID, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	ctx = log.AppendCtx(ctx, slog.String("playlist-id", request.Id.String()))

	// Get Spotify credentials
	s.Env.Logger.DebugContext(ctx, "getting spotify credentials")
	spotifyUserID, accessToken, err := s.getSpotifyCredentials(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "user does not have a spotify integration", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId404JSONResponse{
			Message: "user does not have a spotify integration",
			Status:  apierror.NoSpotifyIntegration.Status(),
			Code:    apierror.NoSpotifyIntegration.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get spotify credentials", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get playlist from database
	playlist, err := s.Env.Database.GetUserPlaylist(ctx, database.GetUserPlaylistParams{
		UserID: userID,
		ID:     request.Id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "playlist not found", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId404JSONResponse{
			Message: "playlist not found",
			Status:  apierror.PlaylistNotFound.Status(),
			Code:    apierror.PlaylistNotFound.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get track URIs
	trackURIs, err := s.getPlaylistTrackURIs(ctx, request.Id)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlist tracks", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Create Spotify playlist
	s.Env.Logger.DebugContext(ctx, "creating spotify playlist")
	spotifyPlaylist, err := s.createSpotifyPlaylist(ctx, accessToken, spotifyUserID, playlist.Name)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create spotify playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	ctx = log.AppendCtx(ctx, slog.String("spotify-playlist-id", spotifyPlaylist.ID))

	// Add tracks to playlist
	s.Env.Logger.DebugContext(ctx, "adding tracks to spotify playlist")
	if err = s.addTracksToSpotifyPlaylist(ctx, accessToken, spotifyPlaylist.ID, trackURIs); err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to add tracks to spotify playlist", slog.Any("error", err))
		return PostApiIntegrationsSpotifyPlaylistId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return PostApiIntegrationsSpotifyPlaylistId201JSONResponse{
		Id:  spotifyPlaylist.ID,
		Url: spotifyPlaylist.URL,
	}, nil
}
