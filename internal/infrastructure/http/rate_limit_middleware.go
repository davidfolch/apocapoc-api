package http

import (
	"net/http"
	"strconv"
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
