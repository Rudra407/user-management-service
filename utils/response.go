package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standard API response
type Response struct {
	Status     string      `json:"status"`
	RequestID  string      `json:"request_id"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Errors     []string    `json:"errors,omitempty"`
	PageInfo   *PageInfo   `json:"page_info,omitempty"`
	TotalCount int64       `json:"total_count,omitempty"`
}

// PageInfo represents pagination information
type PageInfo struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"per_page"`
	TotalPage int64 `json:"total_page"`
}

// SuccessResponse returns a success response
func SuccessResponse(c echo.Context, data interface{}, message string) error {
	requestID := c.Request().Header.Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = c.Response().Header().Get(echo.HeaderXRequestID)
	}

	return c.JSON(http.StatusOK, Response{
		Status:    "success",
		RequestID: requestID,
		Message:   message,
		Data:      data,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c echo.Context, statusCode int, message string, errors []string) error {
	requestID := c.Request().Header.Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = c.Response().Header().Get(echo.HeaderXRequestID)
	}

	return c.JSON(statusCode, Response{
		Status:    "error",
		RequestID: requestID,
		Message:   message,
		Errors:    errors,
	})
}

// ValidationErrorResponse returns a validation error response
func ValidationErrorResponse(c echo.Context, message string, errors []string) error {
	return ErrorResponse(c, http.StatusBadRequest, message, errors)
}

// UnauthorizedErrorResponse returns an unauthorized error response
func UnauthorizedErrorResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// ForbiddenErrorResponse returns a forbidden error response
func ForbiddenErrorResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, message, nil)
}

// NotFoundErrorResponse returns a not found error response
func NotFoundErrorResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusNotFound, message, nil)
}

// InternalServerErrorResponse returns an internal server error response
func InternalServerErrorResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, message, nil)
}
