package service

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPStatus represents an HTTP status.
type HTTPStatus struct {
	// Code is the HTTP status code.
	Code int `json:"code"`
	// Message is the HTTP status message.
	Message string `json:"message"`
	// Title is the HTTP status title.
	Title string `json:"title"`
}

const (
	// MessageServiceHealthy is the message for a healthy service.
	MessageServiceHealthy = "ServiceHealthy"
	// MessageUnknownEndpoint is the message for an unknown endpoint.
	MessageUnknownEndpoint = "UnknownEndpoint"
	// MessageUnexpectedError is the message for an unexpected error.
	MessageUnexpectedError = "UnexpectedError"
)

// NewHTTPStatus creates a new HTTP status.
func NewHTTPStatus(message string) *HTTPStatus {
	messageCodes := map[string]int{
		MessageServiceHealthy:  http.StatusOK,
		MessageUnknownEndpoint: http.StatusNotFound,
		MessageUnexpectedError: http.StatusInternalServerError,
	}

	code, ok := messageCodes[message]
	if !ok {
		code = http.StatusInternalServerError
		message = MessageUnexpectedError
	}

	return &HTTPStatus{
		Code:    code,
		Message: message,
		Title:   http.StatusText(code),
	}
}

// SendStatus sends an HTTP status.
func SendStatus(c echo.Context, message string) error {
	status := NewHTTPStatus(message)

	return c.JSON(status.Code, status)
}

// NewCatchAllHandler creates a new catch-all handler.
func NewCatchAllHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return SendStatus(c, MessageUnknownEndpoint)
	}
}

// NewHealthCheckHandler creates a new health check handler.
func NewHealthCheckHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return SendStatus(c, MessageServiceHealthy)
	}
}
