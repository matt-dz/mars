package openapi

import (
	"bytes"
	"context"

	"mars/docs"
	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
)

func (Server) GetApiOpenapiYaml(
	ctx context.Context, request GetApiOpenapiYamlRequestObject,
) (GetApiOpenapiYamlResponseObject, error) {
	reqid := requestid.FromContext(ctx)

	data, err := docs.Docs.ReadFile("api.yaml")
	if err != nil {
		return GetApiOpenapiYaml500JSONResponse{
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
			Message: "Internal Server Error",
			ErrorId: reqid,
		}, nil
	}

	return GetApiOpenapiYaml200ApplicationxYamlResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}
