package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func TestAiChatDateLoadsDatabaseInsteadOfDashboard(t *testing.T) {
	input := models.AiChatRequest{
		Question: "tanggal 9 oktober 2026 apa mesin yang paling produktif",
		Date:     "2026-10-12", Shift: "CURRENT", Location: "GM1", Status: "GOOD",
		Machines: []models.AiInsightMachineInput{{UUID: "today"}},
		Insight:  &models.AiInsightAnalyzeResponse{},
	}
	calls := 0
	load := func(_ context.Context, date, shift string) (models.APIResponse, error) {
		calls++
		if date != "2026-10-09" || shift != "ALL" {
			t.Fatalf("wrong query: %s %s", date, shift)
		}
		return models.APIResponse{Rows: []models.ProductivityRow{
			{UUID: "historical", NickName: "Historical", Location: "GM1 LINE 1", Status: "GOOD", RuntimeSec: 100, ProcSec: 90, ProductivityPct: 90, MacState: "1"},
			{UUID: "other-area", Location: "GM2 LINE 1", Status: "GOOD"},
			{UUID: "other-status", Location: "GM1 LINE 1", Status: "BAD"},
		}}, nil
	}
	if err := prepareAiChatDateContext(context.Background(), &input, time.Now(), load); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || input.Date != "2026-10-09" || input.DataSource != "database" || input.Insight != nil || len(input.Machines) != 1 || input.Machines[0].UUID != "historical" || input.Machines[0].MacState != "" {
		t.Fatalf("stale or wrong context: %+v", input)
	}
	t.Setenv("AI_PROVIDER", "ollama")
	reply, err := utils.CallOpenAIChat(context.Background(), input)
	if err != nil || reply.Model != "server-rank" || !strings.Contains(reply.Reply, "Historical") || !strings.Contains(reply.Reply, "2026-10-09") || strings.Contains(reply.Reply, "9 mesin") {
		t.Fatalf("incorrect dated answer: %+v, err=%v", reply, err)
	}
}

func TestAiChatDateRejectsUnavailableOrAmbiguousData(t *testing.T) {
	dbErr := errors.New("database unavailable")
	input := models.AiChatRequest{Question: "mesin paling produktif 2026-10-09"}
	err := prepareAiChatDateContext(context.Background(), &input, time.Now(), func(context.Context, string, string) (models.APIResponse, error) { return models.APIResponse{}, dbErr })
	if !errors.Is(err, dbErr) {
		t.Fatalf("error swallowed: %v", err)
	}
	for _, q := range []string{"mesin paling produktif", "tanggal 30 februari 2026 mesin paling produktif", "jam berapa hari ini"} {
		input = models.AiChatRequest{Question: q}
		prepareAiChatDateContext(context.Background(), &input, time.Now(), func(context.Context, string, string) (models.APIResponse, error) {
			t.Fatalf("unexpected database lookup for %s", q)
			return models.APIResponse{}, nil
		})
	}
}

func TestAiChatBestMachinesAndDateFollowupsUseDatabaseRanking(t *testing.T) {
	t.Setenv("AI_PROVIDER", "ollama")
	// An unreachable Ollama URL makes accidental routing to the model fail this test.
	t.Setenv("OLLAMA_BASE_URL", "http://127.0.0.1:1")
	for _, tc := range []struct {
		question  string
		history   []models.AiChatMessage
		wantLower bool
	}{
		{question: "kalo hari kamis tanggal 8 oktober 2026 top mesin terbaik?"},
		{question: "kalo hari kamis tanggal 8 oktober 2026?", history: []models.AiChatMessage{
			{Role: "user", Content: "tanggal 9 oktober 2026 apa mesin yang paling produktif"},
			{Role: "assistant", Content: "Mesin contoh 80%"},
		}},
		{question: "kalau 2026-10-08?", history: []models.AiChatMessage{
			{Role: "user", Content: "top 1 mesin terbaik tanggal 9 Oktober 2026"},
			{Role: "assistant", Content: "Nama lama 80%"},
			{Role: "user", Content: "kalau 10 Oktober 2026?"},
			{Role: "assistant", Content: "Jawaban lain"},
		}},
		{question: "kalau tanggal 8 Oktober 2026?", wantLower: true, history: []models.AiChatMessage{
			{Role: "user", Content: "1 mesin dengan produktivitas terendah pada 2026-10-09"},
		}},
	} {
		t.Run(tc.question, func(t *testing.T) {
			input := models.AiChatRequest{Question: tc.question, Date: "2026-10-12", History: tc.history,
				Machines: []models.AiInsightMachineInput{{UUID: "stale", MachineName: "Dashboard lama", ProductivityPct: 100}},
			}
			calls := 0
			load := func(_ context.Context, date, shift string) (models.APIResponse, error) {
				calls++
				if date != "2026-10-08" {
					t.Fatalf("wrong database date: %s", date)
				}
				return models.APIResponse{Rows: []models.ProductivityRow{
					{UUID: "low", NickName: "Jack Rendah", ProductivityPct: 12.34, RuntimeSec: 100},
					{UUID: "high", NickName: "Jack Tertinggi", ProductivityPct: 96.54, RuntimeSec: 100},
				}}, nil
			}
			if err := prepareAiChatDateContext(context.Background(), &input, time.Now(), load); err != nil {
				t.Fatal(err)
			}
			resp, err := utils.CallOpenAIChat(context.Background(), input)
			if calls != 1 || err != nil || resp.Model != "server-rank" || !strings.Contains(resp.Reply, "2026-10-08") || strings.Contains(resp.Reply, "Dashboard lama") || strings.Contains(resp.Reply, "Mesin contoh") {
				t.Fatalf("wrong routing or stale data: %+v, error=%v, lookups=%d", resp, err, calls)
			}
			if tc.wantLower {
				if strings.Contains(resp.Reply, "Jack Tertinggi") || !strings.Contains(resp.Reply, "12.34%") {
					t.Fatal(resp.Reply)
				}
			} else if !strings.Contains(resp.Reply, "1. Jack Tertinggi") || !strings.Contains(resp.Reply, "96.54%") {
				t.Fatal(resp.Reply)
			}
		})
	}
}

func TestAiChatDatedBestMachinesWithoutActivity(t *testing.T) {
	t.Setenv("AI_PROVIDER", "ollama")
	t.Setenv("OLLAMA_BASE_URL", "http://127.0.0.1:1")
	input := models.AiChatRequest{Question: "top mesin terbaik tanggal 8 Oktober 2026"}
	err := prepareAiChatDateContext(context.Background(), &input, time.Now(), func(context.Context, string, string) (models.APIResponse, error) { return models.APIResponse{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	resp, err := utils.CallOpenAIChat(context.Background(), input)
	if err != nil || !strings.Contains(resp.Reply, "Tidak ada data aktivitas mesin") || resp.Model != "server-rank" {
		t.Fatalf("%+v, %v", resp, err)
	}
}
