package error

import "net/http"

type ErrorCode string

const (
	UnknownError            ErrorCode = "unknown_error"
	InternalServerError     ErrorCode = "internal_server_error"
	BadRequest              ErrorCode = "bad_request"
	UnprocessibleEntity     ErrorCode = "unprocessible_entity"
	InvalidCredentials      ErrorCode = "invalid_credentials"
	InvalidAccessToken      ErrorCode = "invalid_access_token"
	ExpiredAccessToken      ErrorCode = "expired_access_token"
	InvalidRefreshToken     ErrorCode = "invalid_refresh_token"
	ExpiredRefreshToken     ErrorCode = "expired_refresh_token"
	InsufficientPermissions ErrorCode = "insufficient_permissions"
	NoSpotifyIntegration    ErrorCode = "no_spotify_integration"
	NoTracksListened        ErrorCode = "no_tracks_listened"
)

var errorCodeToStatusCode = map[ErrorCode]int{
	UnknownError:            0, // No error code - unknown
	InternalServerError:     http.StatusInternalServerError,
	BadRequest:              http.StatusBadRequest,
	UnprocessibleEntity:     http.StatusUnprocessableEntity,
	InvalidAccessToken:      http.StatusUnauthorized,
	ExpiredAccessToken:      http.StatusUnauthorized,
	InsufficientPermissions: http.StatusForbidden,
	InvalidRefreshToken:     http.StatusUnauthorized,
	ExpiredRefreshToken:     http.StatusUnauthorized,
	InvalidCredentials:      http.StatusUnauthorized,
	NoSpotifyIntegration:    http.StatusNotFound,
	NoTracksListened:        http.StatusConflict,
}

func (ec ErrorCode) Status() int {
	return errorCodeToStatusCode[ec]
}

func (ec ErrorCode) String() string {
	return string(ec)
}
