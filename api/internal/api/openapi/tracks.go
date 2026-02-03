package openapi

import (
	"context"
	"log/slog"
	"time"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/database"
	"mars/internal/log"
	"mars/internal/tokens"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) GetApiTracksTop(
	ctx context.Context, request GetApiTracksTopRequestObject) (
	GetApiTracksTopResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)
	userid, err := tokens.UserIDFromContext(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get userid", slog.Any("error", err))
		return GetApiTracksTop500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	// Compute start and end time
	s.Env.Logger.DebugContext(ctx, "computing start and end times")
	now := time.Now()
	var startTime int64
	var endTime int64

	if request.Params.End != nil {
		endTime = *request.Params.End
	} else {
		endTime = now.Unix()
	}

	if request.Params.Start != nil {
		startTime = *request.Params.Start
	} else {
		startTime = now.Add(-24 * time.Hour).Unix()
	}

	ctx = log.AppendCtx(ctx, slog.Time("start", time.Unix(startTime, 0)))
	ctx = log.AppendCtx(ctx, slog.Time("end", time.Unix(startTime, 0)))
	s.Env.Logger.DebugContext(ctx, "computed")

	if endTime < startTime {
		s.Env.Logger.ErrorContext(ctx, "end time is before start time")
		return GetApiTracksTop400JSONResponse{
			Message: "end should be after start",
			Status:  apierror.BadRequest.Status(),
			Code:    apierror.BadRequest.String(),
			ErrorId: reqid,
		}, nil
	}

	// Get tracks
	s.Env.Logger.DebugContext(ctx, "getting tracks")
	tracks, err := s.Env.Database.TopTracksByUserInRange(ctx, database.TopTracksByUserInRangeParams{
		UserID: userid,
		StartDate: pgtype.Timestamptz{
			Time:  time.Unix(startTime, 0),
			Valid: true,
		},
		EndDate: pgtype.Timestamptz{
			Time:  time.Unix(endTime, 0),
			Valid: true,
		},
	})
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to get tracks", slog.Any("error", err))
		return GetApiTracksTop500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}
	resp := GetApiTracksTop200JSONResponse{
		Tracks: make([]PlaylistTrack, len(tracks)),
	}
	for i, t := range tracks {
		resp.Tracks[i] = PlaylistTrack{
			Id:      t.TrackID,
			Name:    t.Name,
			Artists: t.Artists,
			Href:    t.Href,
			Plays:   int(t.ListenCount),
		}
		if t.ImageUrl.Valid {
			resp.Tracks[i].ImageUrl = &t.ImageUrl.String
		}
	}
	return resp, nil
}
