package api

import (
	"testing"
	"time"
)

func TestMetricPeriods(t *testing.T) {
	cases := []struct {
		name      string
		now       time.Time
		wantDay   string
		wantWeek  string
		wantMonth string
	}{
		{
			name:      "late UTC evening is already next day in Moscow",
			now:       time.Date(2026, 10, 5, 22, 30, 0, 0, time.UTC),
			wantDay:   "2026-10-06T00:00:00+03:00",
			wantWeek:  "2026-09-30T00:00:00+03:00",
			wantMonth: "2026-10-01T00:00:00+03:00",
		},
		{
			name:      "week crosses month boundary",
			now:       time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),
			wantDay:   "2026-03-02T00:00:00+03:00",
			wantWeek:  "2026-02-24T00:00:00+03:00",
			wantMonth: "2026-03-01T00:00:00+03:00",
		},
		{
			name:      "first minutes of a month in Moscow",
			now:       time.Date(2026, 10, 31, 21, 5, 0, 0, time.UTC),
			wantDay:   "2026-11-01T00:00:00+03:00",
			wantWeek:  "2026-10-26T00:00:00+03:00",
			wantMonth: "2026-11-01T00:00:00+03:00",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			day, week, month := metricPeriods(tc.now)
			if got := day.Format(time.RFC3339); got != tc.wantDay {
				t.Errorf("day = %s, want %s", got, tc.wantDay)
			}
			if got := week.Format(time.RFC3339); got != tc.wantWeek {
				t.Errorf("week = %s, want %s", got, tc.wantWeek)
			}
			if got := month.Format(time.RFC3339); got != tc.wantMonth {
				t.Errorf("month = %s, want %s", got, tc.wantMonth)
			}
		})
	}
}
