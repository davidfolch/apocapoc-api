package http

import (
	"net/http"
	"testing"
	"time"

	"apocapoc-api/internal/application/queries"
)

func TestHabitStatsFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := registerAndLogin(t, *ts.Router, "statsuser@example.com", "Password123!")

	habitBody := CreateHabitRequest{
		Name:      "Meditation",
		Type:      "BOOLEAN",
		Frequency: "DAILY",
	}
	rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", habitBody, token)
	var habitResp map[string]string
	decodeResponse(t, rr, &habitResp)
	habitID := habitResp["id"]

	t.Run("Stats for new habit should be zero", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "GET", "/api/v1/stats/habits/"+habitID, nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		var stats queries.HabitStatsDTO
		decodeResponse(t, rr, &stats)

		if stats.TotalCompletions != 0 {
			t.Errorf("Expected 0 total completions, got %d", stats.TotalCompletions)
		}
		if stats.CurrentStreak != 0 {
			t.Errorf("Expected 0 current streak, got %d", stats.CurrentStreak)
		}
		if stats.LongestStreak != 0 {
			t.Errorf("Expected 0 longest streak, got %d", stats.LongestStreak)
		}
	})

	today := time.Now().UTC().Format("2006-01-02")

	t.Run("Stats after marking habit once", func(t *testing.T) {
		markReq := MarkHabitRequest{
			ScheduledDate: today,
		}
		rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits/"+habitID+"/mark", markReq, token)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to mark habit: %d - %s", rr.Code, rr.Body.String())
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/stats/habits/"+habitID, nil, token)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rr.Code)
		}

		var stats queries.HabitStatsDTO
		decodeResponse(t, rr, &stats)

		if stats.TotalCompletions != 1 {
			t.Errorf("Expected 1 total completion, got %d", stats.TotalCompletions)
		}
		if stats.CurrentStreak != 1 {
			t.Errorf("Expected current streak of 1, got %d", stats.CurrentStreak)
		}
		if stats.LongestStreak != 1 {
			t.Errorf("Expected longest streak of 1, got %d", stats.LongestStreak)
		}
	})

	t.Run("Stats after unmarking habit", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "DELETE", "/api/v1/habits/"+habitID+"/entries/"+today, nil, token)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to unmark habit: %d", rr.Code)
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/stats/habits/"+habitID, nil, token)

		var stats queries.HabitStatsDTO
		decodeResponse(t, rr, &stats)

		if stats.TotalCompletions != 0 {
			t.Errorf("Expected 0 total completions after unmark, got %d", stats.TotalCompletions)
		}
		if stats.CurrentStreak != 0 {
			t.Errorf("Expected 0 current streak after unmark, got %d", stats.CurrentStreak)
		}
	})
}

func TestHabitUpdateAffectsStats(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := registerAndLogin(t, *ts.Router, "updatestats@example.com", "Password123!")

	habitBody := CreateHabitRequest{
		Name:      "Running",
		Type:      "BOOLEAN",
		Frequency: "DAILY",
	}
	rr := makeRequest(t, *ts.Router, "POST", "/api/v1/habits", habitBody, token)
	var habitResp map[string]string
	decodeResponse(t, rr, &habitResp)
	habitID := habitResp["id"]

	today := time.Now().UTC().Format("2006-01-02")
	markReq := MarkHabitRequest{
		ScheduledDate: today,
	}
	makeRequest(t, *ts.Router, "POST", "/api/v1/habits/"+habitID+"/mark", markReq, token)

	t.Run("Stats remain after updating habit name", func(t *testing.T) {
		updateReq := UpdateHabitRequest{
			Name: "Morning Running",
		}
		rr := makeRequest(t, *ts.Router, "PUT", "/api/v1/habits/"+habitID, updateReq, token)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to update habit: %d", rr.Code)
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/stats/habits/"+habitID, nil, token)

		var stats queries.HabitStatsDTO
		decodeResponse(t, rr, &stats)

		if stats.TotalCompletions != 1 {
			t.Errorf("Expected stats to persist after update, got %d completions", stats.TotalCompletions)
		}
	})

	t.Run("Stats remain available after archiving habit", func(t *testing.T) {
		rr := makeRequest(t, *ts.Router, "DELETE", "/api/v1/habits/"+habitID, nil, token)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to archive habit: %d", rr.Code)
		}

		rr = makeRequest(t, *ts.Router, "GET", "/api/v1/stats/habits/"+habitID, nil, token)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected stats to remain available for archived habit, got %d", rr.Code)
		}

		var stats queries.HabitStatsDTO
		decodeResponse(t, rr, &stats)

		if stats.TotalCompletions != 1 {
			t.Errorf("Expected stats to persist after archiving, got %d completions", stats.TotalCompletions)
		}
	})
}
