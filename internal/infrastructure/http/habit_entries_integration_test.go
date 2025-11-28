package http

import (
	"net/http"
	"testing"
	"time"
)

func TestHabitEntriesFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := registerAndLogin(t, *ts.Router, "entryuser@example.com", "Password123!")

	habitBody := CreateHabitRequest{
		Name:      "Reading",
		Type:      "BOOLEAN",
		Frequency: "DAILY",
	}
	rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", habitBody, token)
	var habitResp map[string]string
	decodeResponse(t, rr, &habitResp)
	habitID := habitResp["id"]

	today := time.Now().UTC().Format("2006-01-02")

	t.Run("Mark habit as complete", func(t *testing.T) {
		reqBody := MarkHabitRequest{
			ScheduledDate: today,
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits/"+habitID+"/mark", reqBody, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Mark habit twice returns conflict", func(t *testing.T) {
		reqBody := MarkHabitRequest{
			ScheduledDate: today,
		}

		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits/"+habitID+"/mark", reqBody, token)

		if rr.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", rr.Code)
		}
	})

	t.Run("Get habit entries", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+habitID+"/entries?page=1&limit=10", nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		var resp HabitEntriesResponse
		decodeResponse(t, rr, &resp)

		if resp.Total != 1 {
			t.Errorf("Expected 1 entry, got %d", resp.Total)
		}

		if len(resp.Entries) != 1 {
			t.Errorf("Expected 1 entry in array, got %d", len(resp.Entries))
		}
	})

	t.Run("Unmark habit", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "DELETE", "/api/v1/habits/"+habitID+"/entries/"+today, nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+habitID+"/entries?page=1&limit=10", nil, token)
		var resp HabitEntriesResponse
		decodeResponse(t, rr, &resp)

		if resp.Total != 0 {
			t.Errorf("Expected 0 entries after unmark, got %d", resp.Total)
		}
	})

	t.Run("Mark with value for counter habit", func(t *testing.T) {
		counterHabitBody := CreateHabitRequest{
			Name:      "Steps",
			Type:      "COUNTER",
			Frequency: "DAILY",
		}
		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", counterHabitBody, token)
		var counterResp map[string]string
		decodeResponse(t, rr, &counterResp)
		counterHabitID := counterResp["id"]

		value := 10000.0
		reqBody := MarkHabitRequest{
			ScheduledDate: today,
			Value:         &value,
		}

		rr = makeRequest(t, *ts.Router, "POST", "/api/v1/habits/"+counterHabitID+"/mark", reqBody, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/habits/"+counterHabitID+"/entries?page=1&limit=10", nil, token)
		var resp HabitEntriesResponse
		decodeResponse(t, rr, &resp)

		if resp.Entries[0].Value == nil {
			t.Error("Expected value in entry")
		} else if *resp.Entries[0].Value != 10000.0 {
			t.Errorf("Expected value 10000, got %f", *resp.Entries[0].Value)
		}
	})
}
