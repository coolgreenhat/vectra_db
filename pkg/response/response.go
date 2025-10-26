package response

import (
	"encoding/json"
	"net/http"
	"time"

	"vectraDB/pkg/errors"
)

type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Meta struct {
	Total int `json:"total,omitempty"`
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

func Success(w http.ResponseWriter, data interface{}) {
	sendResponse(w, http.StatusOK, &Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	sendResponse(w, http.StatusOK, &Response{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	sendResponse(w, http.StatusCreated, &Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, err error) {
	var appErr *errors.AppError
	if ae, ok := err.(*errors.AppError); ok {
		appErr = ae
	} else {
		// Convert regular error to AppError
		appErr = errors.Wrap(err, http.StatusInternalServerError, "internal server error")
	}
	
	sendResponse(w, appErr.Code, &Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Timestamp: time.Now(),
	})
}

func InternalError(w http.ResponseWriter, err error) {
	sendResponse(w, http.StatusInternalServerError, &Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
			Details: err.Error(),
		},
		Timestamp: time.Now(),
	})
}

func sendResponse(w http.ResponseWriter, statusCode int, response *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
