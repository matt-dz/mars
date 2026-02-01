package openapi

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) PostApiPlaylistsMonthly(
	ctx context.Context, request PostApiPlaylistsMonthlyRequestObject) (
	PostApiPlaylistsMonthlyResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	// Create start and end date
	s.Env.Logger.DebugContext(ctx, "compute start and end dates")
	months := map[string]time.Month{
		"january":   time.January,
		"february":  time.February,
		"march":     time.March,
		"april":     time.April,
		"may":       time.May,
		"june":      time.June,
		"july":      time.July,
		"august":    time.August,
		"september": time.September,
		"october":   time.October,
		"november":  time.November,
		"december":  time.December,
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to load location", slog.Any("error", err))
		return PostApiPlaylistsMonthly500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	startDate := time.Date(request.Body.Year, months[string(request.Body.Month)], 1, 0, 0, 0, 0, loc)
	endDate := startDate.AddDate(0, 1, 0)

	// Query the data
	s.Env.Logger.DebugContext(ctx,
		"querying tracks within range", slog.Group("range", slog.Time("start", startDate), slog.Time("end", endDate)))
	rows, err := s.Env.Database.ListensByTrackInRange(ctx, database.ListensByTrackInRangeParams{
		UserID: request.Body.UserId,
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
		return PostApiPlaylistsMonthly500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	if len(rows) == 0 {
		s.Env.Logger.ErrorContext(ctx, "no tracks listened to in this range - not creating playlist")
		return PostApiPlaylistsMonthly422JSONResponse{
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
		return PostApiPlaylistsMonthly500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := database.New(tx)

	playlistName := fmt.Sprintf("%s %d", request.Body.Month, request.Body.Year)
	playlistID, err := qtx.CreateMonthlyPlaylist(ctx, database.CreateMonthlyPlaylistParams{
		UserID: request.Body.UserId,
		Name:   playlistName,
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to create playlist", slog.Any("error", err))
		return PostApiPlaylistsMonthly500JSONResponse{
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
			return PostApiPlaylistsMonthly500JSONResponse{
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
		return PostApiPlaylistsMonthly500JSONResponse{
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

	return PostApiPlaylistsMonthly201JSONResponse{
		Id: playlistID,
	}, nil
}
