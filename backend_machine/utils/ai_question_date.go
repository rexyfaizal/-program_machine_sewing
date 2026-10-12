package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var aiISODate = regexp.MustCompile(`\b\d{4}-\d{1,2}-\d{1,2}\b`)
var aiNumericDate = regexp.MustCompile(`\b(\d{1,2})[/-](\d{1,2})[/-](\d{4})\b`)
var aiNamedDate = regexp.MustCompile(`(?i)\b(\d{1,2})\s+(januari|februari|maret|april|mei|juni|juli|agustus|september|oktober|november|desember)(?:\s+(\d{4}))?\b`)
var aiIncompleteDate = regexp.MustCompile(`(?i)\btanggal\s+\d`)

func AiQuestionNeedsMachineData(question string) bool {
	ql := strings.ToLower(question)
	for _, term := range []string{"mesin", "produktif", "produktivitas", "produksi", "ranking", "peringkat", "power", "runtime", "durasi", "output", "alarm", "loss", "operator", "proses"} {
		if strings.Contains(ql, term) {
			return true
		}
	}
	return false
}

// ExtractAiQuestionDate recognizes a single requested date, independent of dashboard filters.
// Ambiguous or invalid dates must never fall back to today's dashboard data.
func ExtractAiQuestionDate(question string, now time.Time) (string, bool, error) {
	dates := []string{}
	for _, value := range aiISODate.FindAllString(question, -1) {
		day, err := time.Parse("2006-1-2", value)
		if err != nil {
			return "", true, fmt.Errorf("tanggal %s tidak valid", value)
		}
		dates = append(dates, day.Format("2006-01-02"))
	}
	for _, match := range aiNumericDate.FindAllStringSubmatch(question, -1) {
		day, err := time.Parse("2-1-2006", match[1]+"-"+match[2]+"-"+match[3])
		if err != nil {
			return "", true, fmt.Errorf("tanggal %s tidak valid (gunakan DD/MM/YYYY)", match[0])
		}
		dates = append(dates, day.Format("2006-01-02"))
	}
	months := map[string]int{"januari": 1, "februari": 2, "maret": 3, "april": 4, "mei": 5, "juni": 6, "juli": 7, "agustus": 8, "september": 9, "oktober": 10, "november": 11, "desember": 12}
	for _, match := range aiNamedDate.FindAllStringSubmatch(question, -1) {
		if match[3] == "" {
			return "", true, fmt.Errorf("sertakan tahun tanggal yang diminta, contoh 9 Oktober 2026")
		}
		d, _ := strconv.Atoi(match[1])
		value := fmt.Sprintf("%s-%02d-%02d", match[3], months[strings.ToLower(match[2])], d)
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return "", true, fmt.Errorf("tanggal %s tidak valid", match[0])
		}
		dates = append(dates, value)
	}
	ql := strings.ToLower(question)
	localNow := now.In(jakartaLocation())
	for _, relative := range []struct {
		text   string
		offset int
	}{{"hari ini", 0}, {"kemarin", -1}} {
		if strings.Contains(ql, relative.text) {
			dates = append(dates, localNow.AddDate(0, 0, relative.offset).Format("2006-01-02"))
		}
	}
	if len(dates) > 1 {
		return "", true, fmt.Errorf("chat saat ini mendukung satu tanggal per pertanyaan; gunakan API range untuk rentang tanggal")
	}
	if len(dates) == 1 {
		return dates[0], true, nil
	}
	if aiIncompleteDate.MatchString(question) {
		return "", true, fmt.Errorf("sertakan tanggal lengkap, contoh 9 Oktober 2026 atau 2026-10-09")
	}
	return "", false, nil
}
