package openapi

import (
	"context"
	"log/slog"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
)

func (s Server) GetHealth(ctx context.Context, request GetHealthRequestObject) (GetHealthResponseObject, error) {
	reqid := requestid.FromContext(ctx)

	err := s.Env.Database.Ping(ctx)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to ping database", slog.Any("error", err))
		return GetHealth500JSONResponse{
			ErrorId: reqid,
			Code:    apierror.InternalServerError.String(),
			Message: "Internal Server Error",
			Status:  apierror.InternalServerError.Status(),
		}, nil
	}

	return GetHealth204Response{}, nil
}
