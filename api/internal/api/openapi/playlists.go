package openapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"
	"mars/internal/tokens"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) PostApiPlaylists(
	ctx context.Context, request PostApiPlaylistsRequestObject) (
	PostApiPlaylistsResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	// Create start and end date
	s.Env.Logger.DebugContext(ctx, "compute start and end dates")
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to load location", slog.Any("error", err))
		return PostApiPlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	var startDate time.Time
	var endDate time.Time
	var userid uuid.UUID
	var playlistType string
	if req, err := request.Body.AsWeeklyOrMonthlyRequest(); err == nil {
		userid = req.UserId
		playlistType = string(req.Type)
		switch req.Type {
		case "weekly":
			startDate = time.Date(req.StartDate.Year, time.Month(req.StartDate.Month), req.StartDate.Day, 0, 0, 0, 0, loc)
			endDate = startDate.AddDate(0, 0, 7)
		case "monthly":
			startDate = time.Date(req.StartDate.Year, time.Month(req.StartDate.Month), 1, 0, 0, 0, 0, loc)
			endDate = startDate.AddDate(0, 1, 0)
		}
	}
	if req, err := request.Body.AsCustomRequest(); err == nil && req.Type == "custom" {
		playlistType = string(req.Type)
		userid = req.UserId
		startDate = time.Date(req.StartDate.Year, time.Month(req.StartDate.Month), req.StartDate.Day, 0, 0, 0, 0, loc)
		endDate = time.Date(req.EndDate.Year, time.Month(req.EndDate.Month), req.EndDate.Day, 0, 0, 0, 0, loc)
	}

	// Query the data
	s.Env.Logger.DebugContext(ctx,
		"querying tracks within range", slog.Group("range", slog.Time("start", startDate), slog.Time("end", endDate)))
	rows, err := s.Env.Database.TopTrackIDsByUserInRange(ctx, database.TopTrackIDsByUserInRangeParams{
		UserID: userid,
		StartDate: pgtype.Timestamptz{
			Time:  startDate,
			Valid: true,
		},
		EndDate: pgtype.Timestamptz{
			Time:  endDate,
			Valid: true,
		},
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to query tracks within range", slog.Any("error", err))
		return PostApiPlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if len(rows) == 0 {
		s.Env.Logger.ErrorContext(ctx, "no tracks listened to in this range - not creating playlist")
		return PostApiPlaylists409JSONResponse{
			Message: "no tracks listened to in this range",
			Status:  apierror.NoTracksListened.Status(),
			Code:    apierror.NoTracksListened.String(),
			ErrorId: reqid,
		}, nil
	}

	// Create playlist and add tracks in a transaction
	s.Env.Logger.DebugContext(ctx, "create playlist and add tracks")
	tx, err := s.Env.Pool.Begin(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to begin transaction", slog.Any("error", err))
		return PostApiPlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := database.New(tx)

	playlistName := fmt.Sprintf("%s %d, %d",
		strings.ToLower(startDate.Month().String())[:3],
		startDate.Day(),
		startDate.Year())
	playlistID, err := qtx.CreatePlaylist(ctx, database.CreatePlaylistParams{
		UserID:       userid,
		Name:         playlistName,
		PlaylistType: database.PlaylistType(playlistType),
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create playlist", slog.Any("error", err))
		return PostApiPlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Add tracks to the playlist
	for _, row := range rows {
		err := qtx.AddPlaylistTrack(ctx, database.AddPlaylistTrackParams{
			PlaylistID: playlistID,
			TrackID:    row.TrackID,
			Plays:      int32(row.ListenCount),
		})
		if err != nil {
			s.Env.Logger.ErrorContext(ctx, "failed to add track to playlist",
				slog.Any("error", err),
				slog.String("track_id", row.TrackID),
			)
			return PostApiPlaylists500JSONResponse{
				Message: "internal server error",
				Status:  apierror.InternalServerError.Status(),
				Code:    apierror.InternalServerError.String(),
				ErrorId: reqid,
			}, nil
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to commit transaction", slog.Any("error", err))
		return PostApiPlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	s.Env.Logger.InfoContext(ctx, "created monthly playlist",
		slog.String("playlist_id", playlistID.String()),
		slog.String("name", playlistName),
		slog.Int("track_count", len(rows)),
	)

	return PostApiPlaylists201JSONResponse{
		Id: playlistID,
	}, nil
}

func (s Server) GetApiMePlaylists(
	ctx context.Context, request GetApiMePlaylistsRequestObject) (
	GetApiMePlaylistsResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userid, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return GetApiMePlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get playlists
	s.Env.Logger.DebugContext(ctx, "getting user playlists")
	playlists, err := s.Env.Database.GetUserPlaylists(ctx, userid)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlists", slog.Any("error", err))
		return GetApiMePlaylists500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	resp := make([]ListPlaylistItem, len(playlists))
	for i, pl := range playlists {
		resp[i] = ListPlaylistItem{
			CreatedAt: pl.CreatedAt.Time,
			Id:        pl.ID,
			Name:      pl.Name,
			Type:      string(pl.PlaylistType),
		}
	}
	return GetApiMePlaylists200JSONResponse{Playlists: resp}, nil
}

func (s Server) GetApiPlaylistsId(
	ctx context.Context, request GetApiPlaylistsIdRequestObject) (
	GetApiPlaylistsIdResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userid, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return GetApiPlaylistsId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get playlist and tracks
	s.Env.Logger.DebugContext(ctx, "getting playlist")
	playlist, err := s.Env.Database.GetUserPlaylist(ctx, database.GetUserPlaylistParams{
		UserID: userid,
		ID:     request.Id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		s.Env.Logger.ErrorContext(ctx, "no rows returned - playlist does not exist", slog.Any("error", err))
		return GetApiPlaylistsId404JSONResponse{
			Message: "playlist not found",
			Status:  apierror.PlaylistNotFound.Status(),
			Code:    apierror.PlaylistNotFound.String(),
			ErrorId: reqid,
		}, nil
	} else if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get user playlist", slog.Any("error", err))
		return GetApiPlaylistsId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	tracks, err := s.Env.Database.GetPlaylistTracks(ctx, playlist.ID)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get playlist tracks", slog.Any("error", err))
		return GetApiPlaylistsId500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Return response
	res := GetApiPlaylistsId200JSONResponse{
		Id:        playlist.ID,
		Name:      playlist.Name,
		Type:      string(playlist.PlaylistType),
		Tracks:    make([]PlaylistTrack, len(tracks)),
		CreatedAt: playlist.CreatedAt.Time,
	}
	for i, t := range tracks {
		res.Tracks[i] = PlaylistTrack{
			Artists: t.Artists,
			Id:      t.ID,
			Name:    t.Name,
			Href:    t.Href,
			Plays:   int(t.Plays),
		}
		if t.ImageUrl.Valid {
			res.Tracks[i].ImageUrl = &t.ImageUrl.String
		}
	}

	return res, nil
}
