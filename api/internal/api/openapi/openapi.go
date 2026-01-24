package openapi

import (
	"bytes"
	"context"

	"mars/docs"
	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
)

func (Server) GetOpenapiYaml(
	ctx context.Context, request GetOpenapiYamlRequestObject,
) (GetOpenapiYamlResponseObject, error) {
	reqid := requestid.FromContext(ctx)

	data, err := docs.Docs.ReadFile("api.yaml")
	if err != nil {
		return GetOpenapiYaml500JSONResponse{
			Code:    apierror.InternalServerError.String(),
			Status:  apierror.InternalServerError.Status(),
			Message: "Internal Server Error",
			ErrorId: reqid,
		}, nil
	}

	return GetOpenapiYaml200ApplicationxYamlResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}
