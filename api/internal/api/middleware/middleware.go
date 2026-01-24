// Package middleware contains middleware functions for the API
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/env"
	"mars/internal/log"

	"github.com/go-chi/httplog/v3"

	oapimw "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oklog/ulid/v2"
)

type requestIDKeyType struct{}

var requestIDKey requestIDKeyType

type Middleware struct {
	Env *env.Env
}

func NewMiddleware(env *env.Env) Middleware {
	return Middleware{
		Env: env,
	}
}

func (m Middleware) LogRequest() func(http.Handler) http.Handler {
	return httplog.RequestLogger(m.Env.Logger, &httplog.Options{
		LogExtraAttrs: func(r *http.Request, reqBody string, respStatus int) []slog.Attr {
			requestID := r.Context().Value(requestIDKey)
			if id, ok := requestID.(uint64); ok {
				return []slog.Attr{slog.Uint64("log_id", id)}
			}
			return []slog.Attr{slog.String("log_id", "N/A")}
		},
	})
}

// AddRequestID adds a request ID to the request context.
func (m Middleware) AddRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqid := ulid.Now()
		r = r.WithContext(log.AppendCtx(r.Context(), slog.Uint64("log_id", reqid)))
		r = r.WithContext(requestid.WithContext(r.Context(), reqid))
		next.ServeHTTP(w, r)
	})
}

// Recoverer recovers from panics and returns a standardized error response.
func (m Middleware) Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if err, ok := rvr.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rvr)
				}

				e := env.FromContext(r.Context())
				reqid := requestid.FromContext(r.Context())

				e.Logger.ErrorContext(r.Context(),
					"panic recovered",
					slog.Any("panic", rvr),
					slog.String("stack", string(debug.Stack())))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(&apierror.Error{
					Code:    apierror.InternalServerError,
					Status:  http.StatusInternalServerError,
					Message: "internal server error",
					ErrorID: reqid,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// OAPIErrorHandler handles errors from oapi-codegen middleware and formats them
// according to your error schema.
func OAPIErrorHandler(
	ctx context.Context,
	err error,
	w http.ResponseWriter,
	r *http.Request,
	opts oapimw.ErrorHandlerOpts,
) {
	// Several scenarios where we are handling an error:
	//   1. An error was returned as an apierror in auth middleware
	//   2. There was a validation error (400-level status)
	//   3. There was an internal server error

	reqid := requestid.FromContext(r.Context())

	// 1. Error was returned from middleware
	var errBody *apierror.Error
	if errors.As(err, &errBody) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(opts.StatusCode)
		_ = json.NewEncoder(w).Encode(errBody) //nolint:errchkjson
		return
	}

	// 2. Validation error (use the status code from opts)
	if opts.StatusCode >= 400 && opts.StatusCode < 500 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(opts.StatusCode)
		_ = json.NewEncoder(w).Encode(&apierror.Error{ //nolint:errchkjson
			Code:    apierror.BadRequest,
			Status:  opts.StatusCode,
			Message: err.Error(),
			ErrorID: reqid,
		})
		return
	}

	// 3. An internal server error was surfaced
	_ = apierror.EncodeInternalError(w, reqid)
}
