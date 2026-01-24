package openapi

import (
	"context"
	"log/slog"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
)

func (s Server) GetApiHealth(ctx context.Context, request GetApiHealthRequestObject) (
	GetApiHealthResponseObject, error,
) {
	reqid := requestid.FromContext(ctx)

	err := s.Env.Database.Ping(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to ping database", slog.Any("error", err))
		return GetApiHealth500JSONResponse{
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Message: "Internal Server Error",
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	return GetApiHealth204Response{}, nil
}
