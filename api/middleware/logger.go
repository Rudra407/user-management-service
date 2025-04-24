package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/user/user-management-service/utils"
)

// RequestLogger creates a middleware that logs all requests
func RequestLogger(logger *utils.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Generate request ID if not already set
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
				c.Request().Header.Set(echo.HeaderXRequestID, reqID)
			}

			// Set response header
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)

			// Create request logger
			reqLogger := logger.WithRequestID(reqID)

			// Log request
			reqLogger.WithFields(map[string]interface{}{
				"method":    c.Request().Method,
				"uri":       c.Request().RequestURI,
				"remote_ip": c.RealIP(),
			}).Info("Request received")

			// Process request
			err := next(c)

			// Log response
			latency := time.Since(start)
			status := c.Response().Status

			fields := map[string]interface{}{
				"status":     status,
				"latency_ms": latency.Milliseconds(),
				"latency":    latency.String(),
				"method":     c.Request().Method,
				"uri":        c.Request().RequestURI,
				"remote_ip":  c.RealIP(),
				"user_agent": c.Request().UserAgent(),
				"bytes_in":   c.Request().ContentLength,
				"bytes_out":  c.Response().Size,
			}

			// Determine log level based on status code
			switch {
			case status >= 500:
				reqLogger.WithFields(fields).Error("Server error")
			case status >= 400:
				reqLogger.WithFields(fields).Warn("Client error")
			case status >= 300:
				reqLogger.WithFields(fields).Info("Redirection")
			default:
				reqLogger.WithFields(fields).Info("Success")
			}

			return err
		}
	}
}
