package openapi

import (
	"context"
	"log/slog"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
)

func (s Server) GetApiUsers(ctx context.Context, request GetApiUsersRequestObject) (GetApiUsersResponseObject, error) {
	reqid := requestid.FromContext(ctx)

	// List users
	s.Env.Logger.DebugContext(ctx, "listing users")
	var limit int32 = 10
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	users, err := s.Env.Database.GetUserIDs(ctx, limit)
	if err != nil {
		s.Env.Logger.ErrorContext(ctx, "failed to list users", slog.Any("error", err))
		return GetApiUsers500JSONResponse{
			Message: "internal server error",
			Status:  apierror.InternalServerError.Status(),
			Code:    apierror.InternalServerError.String(),
			ErrorId: reqid,
		}, nil
	}

	return GetApiUsers200JSONResponse{
		Ids: &users,
	}, nil
}
