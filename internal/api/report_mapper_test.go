package api

import (
	"encoding/json"
	"testing"
)

func testReport(t *testing.T) map[string]any {
	t.Helper()
	notes, _ := json.Marshal(map[string]any{
		"overall_score":   0,
		"algorithm_score": 21,
		"coding_score":    0,
		"comments":        "ok",
		"answer_reviews": []map[string]any{
			{"step_id": "q1", "item_type": "question", "prompt": "Что такое горутина?", "is_correct": false},
			{"step_id": "q2", "item_type": "question", "prompt": "Верный", "is_correct": true},
			{"step_id": "t1", "item_type": "task", "prompt": "Задача", "user_answer": "Нет ответа"},
		},
	})
	return map[string]any{"id": "r1", "notes": string(notes), "weaknesses": "слабо"}
}

func TestMapReportHidesWeakPointsForBasic(t *testing.T) {
	mapped := mapReportResponse(testReport(t), false)
	if mapped["locked"] != true {
		t.Fatal("basic report must be locked")
	}
	if got := mapped["weak_points_count"]; got != 2 {
		t.Fatalf("weak_points_count = %v, want 2", got)
	}
	if len(mapped["weak_points"].([]reportWeakPoint)) != 0 || len(mapped["answer_reviews"].([]reportAnswerReview)) != 0 {
		t.Fatal("basic report must not expose weak points or answer reviews")
	}
	if mapped["weaknesses"] != "" {
		t.Fatal("basic report must not expose weaknesses")
	}
	if mapped["overall_score"] != 11 {
		t.Fatalf("overall_score = %v, want real mean 11", mapped["overall_score"])
	}
}

func TestMapReportShowsWeakPointsForPro(t *testing.T) {
	mapped := mapReportResponse(testReport(t), true)
	points := mapped["weak_points"].([]reportWeakPoint)
	if mapped["locked"] != false || len(points) != 2 {
		t.Fatalf("pro report: locked=%v points=%d", mapped["locked"], len(points))
	}
	if points[0].SourceURL == "" || points[0].Topic == "" {
		t.Fatalf("weak point without source: %+v", points[0])
	}
}
