package handlers

import (
	"errors"
	"net/http"
)

// HTTPError wraps an underlying error with an HTTP status code.
type HTTPError struct {
	Status int
	Err    error
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return http.StatusText(e.Status)
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error { return e.Err }

// NewHTTPError creates a new HTTPError with the given status and error.
func NewHTTPError(status int, err error) *HTTPError {
	if err == nil {
		err = errors.New(http.StatusText(status))
	}
	return &HTTPError{Status: status, Err: err}
}

// Predefined HTTP errors for common statuses.
var (
	ErrForbidden    = NewHTTPError(http.StatusForbidden, nil)
	ErrUnauthorized = NewHTTPError(http.StatusUnauthorized, nil)
	ErrBadRequest   = NewHTTPError(http.StatusBadRequest, nil)
	ErrNotFound     = NewHTTPError(http.StatusNotFound, nil)

	// ErrLoginRequired indicates that the user must be logged in to access the resource.
	ErrLoginRequired = errors.New("Access denied: please login")
)

// Wrapper helpers for common HTTP errors.
func WrapForbidden(err error) error    { return NewHTTPError(http.StatusForbidden, err) }
func WrapUnauthorized(err error) error { return NewHTTPError(http.StatusUnauthorized, err) }
func WrapBadRequest(err error) error   { return NewHTTPError(http.StatusBadRequest, err) }
func WrapNotFound(err error) error     { return NewHTTPError(http.StatusNotFound, err) }
