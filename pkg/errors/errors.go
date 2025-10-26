package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

var (
	ErrNotFound         = New(http.StatusNotFound, "resource not found")
	ErrInvalidInput     = New(http.StatusBadRequest, "invalid input")
	ErrInternalError    = New(http.StatusInternalServerError, "internal server error")
	ErrUnauthorized     = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden        = New(http.StatusForbidden, "forbidden")
	ErrConflict         = New(http.StatusConflict, "conflict")
	ErrTooManyRequests  = New(http.StatusTooManyRequests, "too many requests")
	ErrServiceUnavailable = New(http.StatusServiceUnavailable, "service unavailable")
)

var (
	ErrVectorNotFound   = New(http.StatusNotFound, "vector not found")
	ErrInvalidVector    = New(http.StatusBadRequest, "invalid vector data")
	ErrVectorExists     = New(http.StatusConflict, "vector already exists")
	ErrEmptyQuery       = New(http.StatusBadRequest, "query cannot be empty")
	ErrInvalidDimension = New(http.StatusBadRequest, "invalid vector dimension")
)

var (
	ErrDocumentNotFound = New(http.StatusNotFound, "document not found")
	ErrInvalidDocument  = New(http.StatusBadRequest, "invalid document data")
	ErrDocumentExists   = New(http.StatusConflict, "document already exists")
)
