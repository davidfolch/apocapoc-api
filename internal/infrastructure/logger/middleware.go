package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		logger := Log.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		ctx = logger.WithContext(ctx)
		r = r.WithContext(ctx)

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		event := logger.Info()
		if rw.status >= 400 && rw.status < 500 {
			event = logger.Warn()
		} else if rw.status >= 500 {
			event = logger.Error()
		}

		event.
			Int("status", rw.status).
			Int("size", rw.size).
			Dur("duration", duration).
			Msg("HTTP request")
	})
}

func FromContext(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &Log
	}
	return logger
}

func AddUserID(ctx context.Context, userID string) context.Context {
	logger := FromContext(ctx)
	updatedLogger := logger.With().Str("user_id", userID).Logger()
	return updatedLogger.WithContext(ctx)
}
