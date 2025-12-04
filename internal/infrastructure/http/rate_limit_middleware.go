package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"apocapoc-api/internal/infrastructure/auth"

	"github.com/go-chi/httprate"
)

func RateLimitByUser(jwtService *auth.JWTService, requestsPerMinute int, duration time.Duration) func(http.Handler) http.Handler {
	limiter := httprate.NewRateLimiter(
		requestsPerMinute,
		duration,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				return r.RemoteAddr, nil
			}
			return "user:" + userID, nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
		}),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))

			limiter.Handler(next).ServeHTTP(w, r)
		})
	}
}

func RateLimitByEmail(requests int, duration time.Duration) func(http.Handler) http.Handler {
	limiter := httprate.NewRateLimiter(
		requests,
		duration,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return r.RemoteAddr, nil
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return r.RemoteAddr, nil
			}

			if email, ok := data["email"].(string); ok && email != "" {
				return "email:" + strings.ToLower(email), nil
			}

			return r.RemoteAddr, nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too many password reset attempts. Please try again later."}`))
		}),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter.Handler(next).ServeHTTP(w, r)
		})
	}
}
