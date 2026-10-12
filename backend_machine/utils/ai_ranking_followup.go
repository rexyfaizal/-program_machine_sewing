package utils

import (
	"fmt"
	"regexp"
	"strings"

	"backend_machine/models"
)

var aiDateWords = regexp.MustCompile(`(?i)\b(hari|tanggal|senin|selasa|rabu|kamis|jumat|sabtu|minggu)\b`)
var aiFollowupPrefix = regexp.MustCompile(`(?i)^(bagaimana kalau|bagaimana dengan|gimana dengan|bagaimana|gimana|kalau|kalo|untuk|pada|di)\b\s*`)

func rankingLimit(question string) int {
	fallback := 1
	if strings.Contains(strings.ToLower(question), "top") {
		fallback = 5
	}
	return parseTopN(question, fallback)
}

func dateOnlyRankingFollowup(question string) bool {
	rest := aiISODate.ReplaceAllString(question, " ")
	rest = aiNumericDate.ReplaceAllString(rest, " ")
	rest = aiNamedDate.ReplaceAllString(rest, " ")
	rest = strings.ReplaceAll(strings.ToLower(rest), "hari ini", " ")
	rest = strings.ReplaceAll(rest, "kemarin", " ")
	rest = aiDateWords.ReplaceAllString(rest, " ")
	rest = strings.Join(strings.Fields(rest), " ")
	rest = aiFollowupPrefix.ReplaceAllString(rest, "")
	return strings.Trim(rest, " ?.,!") == ""
}

// ResolveAiRankingDateFollowup carries only the latest user ranking intent to a new date.
// Assistant text is never used to select a query or to supply machine values.
func ResolveAiRankingDateFollowup(input *models.AiChatRequest, date string) {
	if AiQuestionNeedsMachineData(input.Question) || !dateOnlyRankingFollowup(input.Question) {
		return
	}
	for i := len(input.History) - 1; i >= 0; i-- {
		message := input.History[i]
		if message.Role != "user" {
			continue
		}
		q := strings.TrimSpace(message.Content)
		direction := ""
		if looksLikeTopProductiveQuestion(q) {
			direction = "paling produktif"
		}
		if looksLikeLowProductiveQuestion(q) {
			direction = "dengan produktivitas terendah"
		}
		if direction != "" {
			input.Question = fmt.Sprintf("%d mesin %s pada tanggal %s", rankingLimit(q), direction, date)
			return
		}
		if !dateOnlyRankingFollowup(q) {
			return
		}
	}
}
