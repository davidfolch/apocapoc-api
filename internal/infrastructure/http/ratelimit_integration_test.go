package http

import (
	"net/http"
	"testing"
)

func TestGlobalRateLimiting(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := registerAndLogin(t, *ts.Router, "ratelimit@example.com", "Password123!")

	t.Run("Request within rate limit succeeds", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			rr := makeRequest(t, *ts.Router, "GET", "/api/v1/habits", nil, token)
			if rr.Code == http.StatusTooManyRequests {
				t.Errorf("Request %d hit rate limit unexpectedly", i+1)
				break
			}
		}
	})
}

func TestPasswordResetRateLimiting(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	registerBody := RegisterRequest{
		Email:    "resetlimit@example.com",
		Password: "Password123!",
	}
	makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", registerBody, "")

	t.Run("Email-based rate limit for password reset", func(t *testing.T) {
		resetReq := map[string]string{
			"email": "resetlimit@example.com",
		}

		for i := 0; i < 3; i++ {
			rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/forgot-password", resetReq, "")
			if rr.Code == http.StatusTooManyRequests {
				t.Fatalf("Request %d hit rate limit too early (limit is 3)", i+1)
			}
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/forgot-password", resetReq, "")
		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429 after 4th request, got %d", rr.Code)
		}
	})

	t.Run("Different emails have separate rate limits", func(t *testing.T) {
		registerBody2 := RegisterRequest{
			Email:    "resetlimit2@example.com",
			Password: "Password123!",
		}
		makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", registerBody2, "")

		resetReq := map[string]string{
			"email": "resetlimit2@example.com",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/forgot-password", resetReq, "")
		if rr.Code == http.StatusTooManyRequests {
			t.Error("Different email should not be affected by previous email's rate limit")
		}
	})
}
