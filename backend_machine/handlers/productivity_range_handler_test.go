package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseProductivityRange(t *testing.T) {
	for _, tc := range []struct {
		name, start, end string
		valid            bool
	}{
		{"same day", "2026-10-01", "2026-10-01", true},
		{"31 days", "2026-10-01", "2026-10-31", true},
		{"32 days", "2026-10-01", "2026-11-01", false},
		{"reverse", "2026-10-02", "2026-10-01", false},
		{"missing start", "", "2026-10-01", false},
		{"missing end", "2026-10-01", "", false},
		{"invalid date", "2026-02-30", "2026-03-01", false},
		{"leap day", "2024-02-29", "2024-03-01", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseProductivityRange(tc.start, tc.end)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestProductivityRangeRejectsInvalidRequests(t *testing.T) {
	h := &Handler{}
	for _, tc := range []struct {
		method, url string
		status      int
	}{
		{http.MethodPost, "/api/productivity/range", http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/productivity/range", http.StatusBadRequest},
		{http.MethodGet, "/api/productivity/range?start_date=2026-10-02&end_date=2026-10-01", http.StatusBadRequest},
		{http.MethodGet, "/api/productivity/range?start_date=2026-10-01&end_date=2026-11-01", http.StatusBadRequest},
	} {
		r := httptest.NewRequest(tc.method, tc.url, nil)
		w := httptest.NewRecorder()
		h.ProductivityRange(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s %s: status=%d, want=%d", tc.method, tc.url, w.Code, tc.status)
		}
		if tc.status == http.StatusMethodNotAllowed && w.Header().Get("Allow") != http.MethodGet {
			t.Fatal("missing Allow: GET")
		}
	}
}
