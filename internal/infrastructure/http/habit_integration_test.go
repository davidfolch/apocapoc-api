package http

import (
	"net/http"
	"testing"
)

func TestHabitCRUDFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := registerAndLogin(t, *ts.Router, "habituser@example.com", "Password123!")

	var habitID string

	t.Run("Create habit", func(t *testing.T) {
		reqBody := CreateHabitRequest{
			Name:        "Exercise",
			Description: "Daily workout",
			Type:        "BOOLEAN",
			Frequency:   "DAILY",
			CarryOver:   false,
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", reqBody, token)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]string
		decodeResponse(t, rr, &resp)

		habitID = resp["id"]
		if habitID == "" {
			t.Fatal("Expected habit ID in response")
		}
	})

	t.Run("Create habit without auth", func(t *testing.T) {
		reqBody := CreateHabitRequest{
			Name:      "No Auth",
			Type:      "BOOLEAN",
			Frequency: "DAILY",
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", reqBody, "")

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Get all user habits", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "GET", "/api/v1/habits", nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		var habits []UserHabitResponse
		decodeResponse(t, rr, &habits)

		if len(habits) != 1 {
			t.Errorf("Expected 1 habit, got %d", len(habits))
		}

		if habits[0].Name != "Exercise" {
			t.Errorf("Expected habit name 'Exercise', got '%s'", habits[0].Name)
		}
	})

	t.Run("Get habit by ID", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+habitID, nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		var habit UserHabitResponse
		decodeResponse(t, rr, &habit)

		if habit.Name != "Exercise" {
			t.Errorf("Expected habit name 'Exercise', got '%s'", habit.Name)
		}
	})

	t.Run("Update habit", func(t *testing.T) {
		reqBody := UpdateHabitRequest{
			Name:        "Morning Exercise",
			Description: "Updated description",
			Frequency:   "DAILY",
		}

		rr := makeRequest(t, *ts.Router, "PUT", "/api/v1/habits/"+habitID, reqBody, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+habitID, nil, token)
		var habit UserHabitResponse
		decodeResponse(t, rr, &habit)

		if habit.Name != "Morning Exercise" {
			t.Errorf("Expected updated name 'Morning Exercise', got '%s'", habit.Name)
		}
	})

	t.Run("Archive habit", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "DELETE", "/api/v1/habits/"+habitID, nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/habits", nil, token)
		var habits []UserHabitResponse
		decodeResponse(t, rr, &habits)

		if len(habits) != 0 {
			t.Errorf("Expected 0 active habits after archive, got %d", len(habits))
		}
	})

	t.Run("Create habit with missing type returns field-level error", func(t *testing.T) {
		token := registerAndLogin(t, *ts.Router, "validationuser@example.com", "Password123!")

		reqBody := map[string]any{"name": "No Type"}
		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", reqBody, token)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp ValidationErrorResponse
		decodeResponse(t, rr, &resp)
		if resp.Field != "type" {
			t.Errorf("Expected field 'type', got %q. Body: %s", resp.Field, rr.Body.String())
		}
		if resp.Error == "" {
			t.Errorf("Expected non-empty translated error message. Body: %s", rr.Body.String())
		}
	})

	t.Run("Create weekly habit without specific_days returns field-level error", func(t *testing.T) {
		token := registerAndLogin(t, *ts.Router, "weeklyuser@example.com", "Password123!")

		reqBody := CreateHabitRequest{
			Name:      "Cut nails",
			Type:      "BOOLEAN",
			Frequency: "WEEKLY",
		}
		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", reqBody, token)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var resp ValidationErrorResponse
		decodeResponse(t, rr, &resp)
		if resp.Field != "specific_days" {
			t.Errorf("Expected field 'specific_days', got %q. Body: %s", resp.Field, rr.Body.String())
		}
	})

	t.Run("Access other user's habit", func(t *testing.T) {
		otherToken := registerAndLogin(t, *ts.Router, "otheruser@example.com", "Password123!")

		reqBody := CreateHabitRequest{
			Name:      "Other User Habit",
			Type:      "BOOLEAN",
			Frequency: "DAILY",
		}
		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", reqBody, otherToken)
		var createResp map[string]string
		decodeResponse(t, rr, &createResp)
		otherHabitID := createResp["id"]

		rr2 := makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+otherHabitID, nil, token)

		if rr2.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rr2.Code)
		}
	})
}
