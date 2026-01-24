// Package error provides standardized error handling for HTTP responses.
package error

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	Code    ErrorCode `json:"code"`
	Status  int       `json:"status"`
	Message string    `json:"message"`
	ErrorID uint64    `json:"error_id"`
}

func (e *Error) Error() string {
	data, _ := json.Marshal(e) //nolint:errchkjson
	return string(data)
}

func buildError(code ErrorCode, message string, errorID uint64) Error {
	return Error{
		Code:    code,
		Status:  errorCodeToStatusCode[code],
		Message: message,
		ErrorID: errorID,
	}
}

func EncodeError(w http.ResponseWriter, code ErrorCode, message string, errorid uint64) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCodeToStatusCode[code])

	if err := json.NewEncoder(w).Encode(buildError(code, message, errorid)); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}
	return nil
}

func EncodeUnknownError(w http.ResponseWriter, message string, errorid uint64, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(Error{
		Code:    UnknownError,
		Status:  statusCode,
		Message: message,
		ErrorID: errorid,
	}); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}
	return nil
}

func EncodeInternalError(w http.ResponseWriter, errorid uint64) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCodeToStatusCode[InternalServerError])

	res := buildError(InternalServerError, "Internal Server Error", errorid)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	return nil
}
