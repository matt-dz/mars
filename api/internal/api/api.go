// Package api
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"mars/docs"
	apierror "mars/internal/api/error"
	"mars/internal/api/middleware"
	"mars/internal/api/openapi"
	"mars/internal/api/requestid"
	"mars/internal/env"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	oapimw "github.com/oapi-codegen/nethttp-middleware"
)

func Start(ctx context.Context, port uint16, env *env.Env) error {
	server := openapi.NewServer(env)
	spec, err := docs.Docs.ReadFile("api.yaml")
	if err != nil {
		return fmt.Errorf("reading openapi file: %w", err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(spec)
	if err != nil {
		return fmt.Errorf("creating openapi loader: %w", err)
	}
	swagger.Servers = nil

	router := chi.NewMux()
	m := middleware.NewMiddleware(env)
	router.Use(m.AddRequestID)
	router.Use(m.LogRequest())
	router.Use(m.Recoverer)
	router.Use(oapimw.OapiRequestValidatorWithOptions(swagger, &oapimw.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: m.OAPIAuthFunc,
		},
		ErrorHandlerWithOpts: m.OAPIErrorHandler,
	}))

	strictHandlerOptions := openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			reqid := requestid.FromContext(r.Context())
			_ = apierror.EncodeError(w, apierror.BadRequest, err.Error(), reqid)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			reqid := requestid.FromContext(r.Context())
			_ = apierror.EncodeInternalError(w, reqid)
		},
	}

	handler := openapi.HandlerFromMux(
		openapi.NewStrictHandlerWithOptions(server, nil, strictHandlerOptions),
		router,
	)

	s := &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
	}

	errCh := make(chan error, 1)

	// Start server
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for graceful shutdown or server error
	select {
	case <-ctx.Done():
		const timeoutDuration = 10 * time.Second
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil

	case err := <-errCh:
		return err
	}
}
