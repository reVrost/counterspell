package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// LoggingMiddleware logs both requests and responses with details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		slog.Info("Incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"has_auth_header", r.Header.Get("Authorization") != "",
		)

		// Wrap response writer to capture status and body
		rw := &responseWriter{ResponseWriter: w}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Log response
		duration := time.Since(start)
		slog.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
			"response_size", rw.size,
		)
	})
}

// DebugLoggingMiddleware logs detailed request/response info for debugging
func DebugLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request details
		logEntry := slog.With(
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		logEntry.Info("=== DEBUG: Incoming Request ===")
		logEntry.Info("Headers", "headers", r.Header)

		// Log request body for POST/PUT
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			body := new(bytes.Buffer)
			body.ReadFrom(r.Body)
			r.Body = io.NopCloser(body)

			// Try to parse as JSON for pretty logging
			var jsonBody map[string]any
			if err := json.Unmarshal(body.Bytes(), &jsonBody); err == nil {
				logEntry.Info("Request Body (JSON)", "body", jsonBody)
			} else {
				logEntry.Info("Request Body (Raw)", "body", body.String())
			}

			// Reset body for next handler
			r.Body = io.NopCloser(bytes.NewReader(body.Bytes()))
		}

		// Wrap response writer to capture status and body
		rw := &responseWriter{ResponseWriter: w}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Log response details
		duration := time.Since(start)
		logEntry.Info("=== DEBUG: Response ===",
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
			"response_size", rw.size,
		)

		if rw.body.Len() > 0 {
			// Try to parse as JSON for pretty logging
			var jsonResponse map[string]any
			if err := json.Unmarshal(rw.body.Bytes(), &jsonResponse); err == nil {
				logEntry.Info("Response Body (JSON)", "body", jsonResponse)
			} else {
				logEntry.Info("Response Body (Raw)", "body", rw.body.String())
			}
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and body
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	rw.body.Write(b)
	return size, err
}

// Hijack implements http.Hijacker for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
