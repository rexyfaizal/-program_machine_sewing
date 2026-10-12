package utils

import (
	"testing"

	"backend_machine/models"
)

func TestRankingDateFollowupDoesNotReplaceNewTopics(t *testing.T) {
	for _, q := range []string{"tanggal 8 Oktober 2026 hari apa?", "apakah tanggal 8 Oktober 2026 libur?", "top mesin terbaik tanggal 8 Oktober 2026"} {
		input := models.AiChatRequest{Question: q, History: []models.AiChatMessage{{Role: "user", Content: "top 3 mesin terbaik"}}}
		ResolveAiRankingDateFollowup(&input, "2026-10-08")
		if input.Question != q {
			t.Fatalf("new topic replaced: %s", input.Question)
		}
	}
	input := models.AiChatRequest{Question: "kalau 8 Oktober 2026?", History: []models.AiChatMessage{
		{Role: "user", Content: "top 3 mesin terbaik"},
		{Role: "user", Content: "bagaimana cara export Excel?"},
		{Role: "assistant", Content: "top 5 mesin paling produktif"},
	}}
	ResolveAiRankingDateFollowup(&input, "2026-10-08")
	if input.Question != "kalau 8 Oktober 2026?" {
		t.Fatal("resumed an unrelated topic or trusted assistant ranking")
	}
}
