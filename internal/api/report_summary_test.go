package api

import (
	"strings"
	"testing"
)

func TestReportHeadlinePicksWeakerSection(t *testing.T) {
	if got := reportHeadline(60, 85, true, true); !strings.HasPrefix(got, "Хороший результат") {
		t.Fatalf("60/85 headline = %q", got)
	}
	if got := reportHeadline(40, 85, true, true); !strings.HasPrefix(got, "Сильная практика") {
		t.Fatalf("40/85 headline = %q", got)
	}
	if got := reportHeadline(70, 0, true, false); !strings.HasPrefix(got, "Хорошая теория") {
		t.Fatalf("theory-only headline = %q", got)
	}
}

func TestFocusGroupsByTechnology(t *testing.T) {
	groups := focusGroups([]reportWeakPoint{
		{Technology: "Go", Topic: "горутины"},
		{Technology: "SQL", Topic: "индексы"},
		{Technology: "Go", Topic: "context"},
		{Technology: "Go", Topic: "Горутины"},
	})
	if len(groups) != 2 || groups[0].Title != "Go" || groups[0].Count != 3 {
		t.Fatalf("groups = %+v", groups)
	}
	if len(groups[0].Topics) != 2 {
		t.Fatalf("duplicate topics must merge: %+v", groups[0].Topics)
	}
}

func TestMapReportExposesTotalsButNotFocusForBasic(t *testing.T) {
	mapped := mapReportResponse(testReport(t), false)
	if mapped["theory_total"] != 2 || mapped["task_total"] != 1 || mapped["theory_correct"] != 1 {
		t.Fatalf("totals = %v/%v/%v", mapped["theory_correct"], mapped["theory_total"], mapped["task_total"])
	}
	if len(mapped["focus"].([]reportFocusGroup)) != 0 || mapped["focus_count"].(int) == 0 {
		t.Fatalf("basic focus = %v count=%v", mapped["focus"], mapped["focus_count"])
	}
}
