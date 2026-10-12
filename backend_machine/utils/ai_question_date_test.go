package utils

import (
	"strings"
	"testing"
	"time"

	"backend_machine/models"
)

func TestExtractAiQuestionDate(t *testing.T) {
	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC) // 12 October in WIB
	for _, tc := range []struct {
		q, want            string
		mentioned, invalid bool
	}{
		{"tanggal 9 oktober 2026 apa mesin yang paling produktif", "2026-10-09", true, false},
		{"top 3 mesin tanggal 2026-10-09", "2026-10-09", true, false},
		{"mesin pada 09/10/2026", "2026-10-09", true, false},
		{"mesin pada 9-10-2026", "2026-10-09", true, false},
		{"mesin hari ini", "2026-10-12", true, false},
		{"mesin kemarin", "2026-10-11", true, false},
		{"tanggal 30 februari 2026", "", true, true},
		{"tanggal 2026-02-30", "", true, true},
		{"tanggal 9 Oktober", "", true, true},
		{"tanggal 9", "", true, true},
		{"tanggal 2026-10-01 sampai 2026-10-09", "", true, true},
		{"mesin paling produktif", "", false, false},
	} {
		t.Run(tc.q, func(t *testing.T) {
			got, mentioned, err := ExtractAiQuestionDate(tc.q, now)
			if got != tc.want || mentioned != tc.mentioned || (err != nil) != tc.invalid {
				t.Fatalf("got %q, mentioned=%v, err=%v", got, mentioned, err)
			}
		})
	}
}

func TestRankingUsesRequestedDateNotDayNumber(t *testing.T) {
	for _, tc := range []struct {
		q    string
		want int
	}{
		{"tanggal 9 oktober 2026 apa mesin yang paling produktif", 1},
		{"5 mesin paling produktif tanggal 9 oktober 2026", 5},
		{"tanggal 2026-10-09 top 3", 3},
		{"20 mesin paling produktif", 20},
	} {
		if got := parseTopN(tc.q, 1); got != tc.want {
			t.Fatalf("%s: got %d want %d", tc.q, got, tc.want)
		}
	}
	req := models.AiChatRequest{
		Question: "tanggal 9 oktober 2026 apa mesin yang paling produktif",
		Date:     "2026-10-09", DataSource: "database", Shift: "ALL",
		Machines: []models.AiInsightMachineInput{
			{UUID: "a", MachineName: "Mesin A", ProductivityPct: 30, RuntimeSec: 100},
			{UUID: "b", MachineName: "Mesin B", ProductivityPct: 90, RuntimeSec: 100},
		},
	}
	reply, ok := answerRankingQuestion(req)
	if !ok || !strings.Contains(reply, "Mesin B") || strings.Contains(reply, "Mesin A") || !strings.Contains(reply, "2026-10-09") || !strings.Contains(reply, "data database") || strings.Contains(reply, "saat ini") {
		t.Fatalf("wrong historical ranking: %s", reply)
	}
	if looksLikeClockQuestion("mesin paling produktif hari ini") {
		t.Fatal("ranking intercepted as clock question")
	}
}
