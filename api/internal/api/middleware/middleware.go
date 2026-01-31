// Package middleware contains middleware functions for the API
package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"

	apierror "mars/internal/api/error"
	"mars/internal/api/requestid"
	"mars/internal/env"
	marsjwt "mars/internal/jwt"
	"mars/internal/log"
	"mars/internal/role"
	"mars/internal/tokens"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/httplog/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	oapimw "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oklog/ulid/v2"
)

type requestidKeyType struct{}

var requestidKey requestidKeyType

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
			reqid := r.Context().Value(requestidKey)
			if id, ok := reqid.(uint64); ok {
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
func (m Middleware) OAPIErrorHandler(
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

func (m Middleware) OAPIAuthFunc(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	reqid := requestid.FromContext(ctx)

	if input.SecuritySchemeName == "" {
		return nil
	}

	// Get access token
	var accessToken string
	cookie, err := input.RequestValidationInput.Request.Cookie(tokens.AccessTokenName)
	if err != nil {
		m.Env.Logger.DebugContext(ctx, "no access token cookie, reading header", slog.Any("error", err))
		authHeader := input.RequestValidationInput.Request.Header.Get(tokens.AuthorizationHeader)
		accessToken, err = tokens.ParseBearerToken(authHeader)
		if err != nil {
			m.Env.Logger.ErrorContext(ctx, "failed to parse authorization header", slog.Any("error", err))
			return &apierror.Error{
				Code:    apierror.InvalidAccessToken,
				Status:  apierror.InvalidAccessToken.Status(),
				Message: "access token invalid or missing",
				ErrorID: reqid,
			}
		}
	} else {
		accessToken = cookie.Value
	}

	// Validate CSRF token
	if slices.Contains(
		[]string{http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodDelete},
		input.RequestValidationInput.Request.Method) {
		// State-changing request - validate csrf tokens
		m.Env.Logger.DebugContext(ctx, "validating csrf tokens")
		if err := validateCSRFHeader(input); err != nil {
			m.Env.Logger.ErrorContext(ctx, "failed to validate csrf token", slog.Any("error", err))
			return &apierror.Error{
				Code:    apierror.InvalidCredentials,
				Status:  apierror.InvalidCredentials.Status(),
				Message: err.Error(),
				ErrorID: reqid,
			}
		}
	}

	// Get app secret
	appSecret := m.Env.Get("APP_SECRET")
	if appSecret == "" {
		m.Env.Logger.ErrorContext(ctx, "APP_SECRET not set")
		return &apierror.Error{
			Code:    apierror.InternalServerError,
			Status:  apierror.InternalServerError.Status(),
			Message: "internal server error",
			ErrorID: reqid,
		}
	}
	if err != nil {
		m.Env.Logger.ErrorContext(ctx, "failed to decode app secret", slog.Any("error", err))
		return &apierror.Error{
			Code:    apierror.InternalServerError,
			Status:  apierror.InternalServerError.Status(),
			Message: "internal server error",
			ErrorID: reqid,
		}
	}

	// Validate JWT
	jwtAccess, err := marsjwt.ValidateJWT(accessToken, marsjwt.DefaultKID, []byte(appSecret))
	if errors.Is(err, jwt.ErrTokenExpired) {
		m.Env.Logger.ErrorContext(ctx, "jwt expired", slog.Any("error", err))
		return &apierror.Error{
			Code:    apierror.ExpiredAccessToken,
			Status:  apierror.ExpiredAccessToken.Status(),
			Message: "access token expired",
			ErrorID: reqid,
		}
	} else if err != nil {
		m.Env.Logger.ErrorContext(ctx, "failed to validate jwt", slog.Any("error", err))
		return &apierror.Error{
			Code:    apierror.InvalidAccessToken,
			Status:  apierror.InvalidAccessToken.Status(),
			Message: "invalid access token",
			ErrorID: reqid,
		}
	}

	// Extract user id
	sub, err := jwtAccess.Claims.GetSubject()
	if err != nil {
		m.Env.Logger.ErrorContext(ctx, "failed to get subject", slog.Any("errro", err))
		return &apierror.Error{
			Code:    apierror.InternalServerError,
			Status:  apierror.InternalServerError.Status(),
			Message: "internal server error",
			ErrorID: reqid,
		}
	}
	userid, err := uuid.Parse(sub)
	if err != nil {
		m.Env.Logger.ErrorContext(ctx, "failed to parse sub", slog.Any("error", err))
		return &apierror.Error{
			Code:    apierror.InternalServerError,
			Status:  apierror.InternalServerError.Status(),
			Message: "internal server error",
			ErrorID: reqid,
		}
	}

	// Authorize user
	roleClaim := jwtAccess.Claims.(jwt.MapClaims)["role"].(string)
	userRole := role.ToRole(roleClaim)
	if input.RequestValidationInput.Request.URL.Path == "/api/oauth/spotify/token/refresh" && userRole != role.RoleAdmin {
		return &apierror.Error{
			Code:    apierror.InsufficientPermissions,
			Status:  apierror.InsufficientPermissions.Status(),
			Message: "user does not have admin role",
			ErrorID: reqid,
		}
	}

	// Store user info in context
	r := input.RequestValidationInput.Request
	r = r.WithContext(log.AppendCtx(r.Context(), slog.String("user-id", userid.String())))
	r = r.WithContext(tokens.UserIDWithContext(r.Context(), userid))
	r = r.WithContext(tokens.AccessTokenWithContext(r.Context(), jwtAccess))
	*input.RequestValidationInput.Request = *r

	return nil
}

func validateCSRFHeader(input *openapi3filter.AuthenticationInput) error {
	csrfHeader := input.RequestValidationInput.Request.Header.Get(tokens.CsrfTokenHeader)
	if csrfHeader == "" {
		return fmt.Errorf("missing token header %q", tokens.CsrfTokenHeader)
	}
	csrfCookie, err := input.RequestValidationInput.Request.Cookie(tokens.CsrfTokenName)
	if err != nil {
		return fmt.Errorf("missing token cookie %q", tokens.CsrfTokenName)
	}
	if subtle.ConstantTimeCompare([]byte(csrfCookie.Value), []byte(csrfHeader)) == 0 {
		return errors.New("csrf tokens do not match")
	}

	return nil
}
