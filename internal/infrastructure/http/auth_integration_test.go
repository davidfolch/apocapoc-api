package http

import (
	"net/http"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	t.Run("Register new user", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", reqBody, "")

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp RegisterResponse
		decodeResponse(t, rr, &resp)

		if resp.UserID == "" {
			t.Error("Expected user ID in response")
		}
		if resp.Message == "" {
			t.Error("Expected message in response")
		}
	})

	t.Run("Register duplicate email", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "Password123!",
		}

		makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", reqBody, "")

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", reqBody, "")

		if rr.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", rr.Code)
		}
	})

	t.Run("Register with invalid email", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "invalid-email",
			Password: "Password123!",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", reqBody, "")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Register with short password", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "short@example.com",
			Password: "123",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", reqBody, "")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})

	t.Run("Login with valid credentials", func(t *testing.T) {
		registerBody := RegisterRequest{
			Email:    "login@example.com",
			Password: "Password123!",
		}
		makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", registerBody, "")

		loginBody := LoginRequest{
			Email:    "login@example.com",
			Password: "Password123!",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/login", loginBody, "")

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var resp AuthResponse
		decodeResponse(t, rr, &resp)

		if resp.Token == "" {
			t.Error("Expected token in response")
		}
	})

	t.Run("Login with invalid password", func(t *testing.T) {
		registerBody := RegisterRequest{
			Email:    "wrongpass@example.com",
			Password: "Password123!",
		}
		makeRequest(t, *ts.Router, "POST", "/api/v1/auth/register", registerBody, "")

		loginBody := LoginRequest{
			Email:    "wrongpass@example.com",
			Password: "wrongpassword",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/login", loginBody, "")

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Login with non-existent user", func(t *testing.T) {
		loginBody := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "Password123!",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/auth/login", loginBody, "")

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})
}
