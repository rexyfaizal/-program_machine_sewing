package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"backend_machine/models"
)

const (
	defaultOpenAIModel = "gpt-4o-mini"
	defaultOllamaModel = "llama3.2"
	defaultOllamaURL   = "http://127.0.0.1:11434"
	openAIAPIURL       = "https://api.openai.com/v1/chat/completions"
)

type openAIChatRequest struct {
	Model          string              `json:"model"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat map[string]string   `json:"response_format,omitempty"`
	Messages       []openAIChatMessage `json:"messages"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type openAIStructuredReply struct {
	Reply   string               `json:"reply"`
	Actions []models.AiChatAction `json:"actions"`
}

// AiChatProvider openai | ollama (default ollama).
func AiChatProvider() string {
	p := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	if p == "" {
		return "ollama"
	}
	return p
}

// AiChatConfigured true jika provider siap dikonfigurasi.
func AiChatConfigured() bool {
	switch AiChatProvider() {
	case "ollama":
		return true
	case "openai":
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != ""
	default:
		return false
	}
}

// OpenAIConfigured kompatibilitas lama.
func OpenAIConfigured() bool {
	return AiChatConfigured()
}

// AiChatModel nama model aktif.
func AiChatModel() string {
	switch AiChatProvider() {
	case "ollama":
		model := strings.TrimSpace(os.Getenv("OLLAMA_MODEL"))
		if model == "" {
			return defaultOllamaModel
		}
		return model
	default:
		model := strings.TrimSpace(os.Getenv("OPENAI_MODEL"))
		if model == "" {
			return defaultOpenAIModel
		}
		return model
	}
}

// OpenAIModel kompatibilitas lama.
func OpenAIModel() string {
	return AiChatModel()
}

func ollamaBaseURL() string {
	base := strings.TrimSpace(os.Getenv("OLLAMA_BASE_URL"))
	if base == "" {
		base = defaultOllamaURL
	}
	return strings.TrimRight(base, "/")
}

func aiChatCompletionsURL() string {
	if AiChatProvider() == "ollama" {
		return ollamaBaseURL() + "/v1/chat/completions"
	}
	return openAIAPIURL
}

// BuildAiChatSystemPrompt instruksi Copilot dashboard.
func BuildAiChatSystemPrompt() string {
	return strings.TrimSpace(`
Kamu adalah Copilot produksi mesin jahit (GGI / Xinghuo) di dashboard internal pabrik.
Bahasa: Indonesia, profesional, ringkas, seperti asisten keputusan utilisasi mesin — bukan chatbot santai.

Menu aplikasi:
1) dashboard = Dashboard Produktivitas (KPI, ranking, AI Insight, tabel mesin, filter area/shift/status/tanggal)
2) detail = Detail Proses mesin (cycle/hour bar)
3) location = Template Lokasi / pengaturan line & shift
4) master-ie = Master IE (style-proses, CT GM3)
5) operator-productivity = Produktivitas Operator (CT, target, koreksi style)
6) mechanic = Monitoring mekanik (jangan dorong tiket kecuali user minta)

Aturan ketat:
- JANGAN mengarang angka produksi. Pakai hanya data di konteks JSON (machines + insightItems).
- Utilisasi mesin = productivityPct/utilPct di JSON — sama dengan dashboard. Jangan hitung ulang dengan rumus lain.
- Pertanyaan hari/tanggal/jam SAAT INI: jawab dari field now (nowDate, nowTime, dayNameId, timezone). JANGAN bilang "tidak ada info tanggal" jika field now ada.
- Field date = tanggal filter dashboard (bisa beda dari hari ini). Jelaskan bedanya jika user tanya.
- Pertanyaan jumlah/total mesin di dashboard: pakai machineCountTotal (semua mesin di filter). JANGAN pakai panjang array machines saja (bisa terpotong) atau hanya mesin aktif.
- Pertanyaan mesin sedang ON/nyala/online: pakai machineCountOn (macState 1+2). Boleh sebut rincian working/online/offline.
- machineCountActive = mesin yang punya Power On/Process/produktivitas terukur hari ini. Beda dari total dan beda dari ON.
- Pertanyaan ranking/produktif (contoh: "5 mesin paling produktif"): urutkan machines by utilPct/productivityPct. JANGAN jawab dengan daftar Insight HIGH / prioritas intervensi.
- JANGAN pakai output/pcs sebagai acuan utama (data sering tidak akurat).
- JANGAN mengandalkan antrian tiket mekanik sebagai sumber keputusan.
- Boleh jelaskan arti menu/filter/status.
- Analisa mesin dari: power on (runtimeSec), process (procSec), lossTimeSec, macState, status, operator, note, dan blok insight (bukti rule).
- Jika data kurang, katakan kurang — JANGAN spekulasi penyebab (hanca/material/rusak/operator malas) kecuali lossReasonCode/note/insight evidence menyebutnya.
- Kode IDLE_NO_NOTE = celah data (belum ada keterangan). Aksi yang benar: minta isi keterangan. DILARANG menebak tunggu hanca/mesin rusak.
- Kode FREQUENT_BROKEN / field brokenCount7d = mesin sering rusak (riwayat tercatat). Sarankan preventive: cek teknis sebelum shift berikutnya, spare/mekanik standby — jangan tebak jenis kerusakan.
- Jawaban harus spesifik per mesin: sebut nama/line + util% + power on/process dari data. Dilarang template teks sama untuk semua mesin.
- Pertanyaan "apa yang harus dilakukan 1 jam ke depan / prioritas": WAJIB jawab dari insightItems (decision + improvement + evidence). Format bernomor, sebut line/mesin + aksi konkret + angka.
- DILARANG jawaban generik seperti "mungkin periksa mesin BAD", "tidak ada informasi jelas", "lakukan perbaikan jika diperlukan" tanpa menunjuk item insight.
- Aksi yang mengubah UI HARUS butuh konfirmasi user (needsConfirm=true).
- Jangan usulkan eksekusi berbahaya (ubah shift setting, logout massal, hapus data).

Output WAJIB JSON valid:
{
  "reply": "penjelasan untuk user (markdown singkat boleh)",
  "actions": [
    {
      "type": "set_location|set_shift|set_status|set_date|navigate|refresh_insight|open_chart",
      "value": "nilai aksi",
      "label": "teks tombol singkat",
      "needsConfirm": true
    }
  ]
}

Panduan actions:
- set_location: GM1/GM2/GM3/ALL atau nama line
- set_shift: SHIFT_1|SHIFT_2|SHIFT_3|CURRENT|NORMAL|ALL
- set_status: ALL|GOOD|NORMAL|BAD
- set_date: YYYY-MM-DD
- navigate: dashboard|detail|location|master-ie|operator-productivity
- refresh_insight: "1"
- open_chart: UUID mesin (wajib UUID dari konteks machines) untuk membuka chart produktivitas mesin di dashboard
Kosongkan actions jika tidak perlu.
Jika user minta grafik/chart/diagram mesin tertentu, sertakan action open_chart dengan uuid mesin yang cocok.
`)
}

type aiClockInfo struct {
	Date      string `json:"nowDate"`
	Time      string `json:"nowTime"`
	DateTime  string `json:"nowDateTime"`
	DayNameId string `json:"dayNameId"`
	Timezone  string `json:"timezone"`
}

func jakartaLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}

func indonesianDayName(t time.Time) string {
	names := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	return names[int(t.Weekday())]
}

func currentAiClock() aiClockInfo {
	now := time.Now().In(jakartaLocation())
	return aiClockInfo{
		Date:      now.Format("2006-01-02"),
		Time:      now.Format("15:04:05"),
		DateTime:  now.Format("2006-01-02 15:04:05"),
		DayNameId: indonesianDayName(now),
		Timezone:  "Asia/Jakarta (WIB)",
	}
}

func looksLikeClockQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	keys := []string{
		"hari apa", "hari ini", "tanggal berapa", "tanggal hari", "tanggal sekarang",
		"sekarang tanggal", "jam berapa", "jam sekarang", "waktu sekarang",
		"hari dan tanggal", "tanggal dan hari", "tanggalnya",
	}
	for _, k := range keys {
		if strings.Contains(ql, k) {
			return true
		}
	}
	return false
}

type machineCountStats struct {
	Total    int
	Active   int // punya jejak runtime/proc/produktivitas di filter tanggal
	On       int // macState 1 (online) atau 2 (working)
	Working  int // macState 2
	Online   int // macState 1
	Offline  int // macState 0 / kosong
}

func countMachinesInRequest(machines []models.AiInsightMachineInput) machineCountStats {
	stats := machineCountStats{}
	seen := map[string]bool{}
	for _, m := range machines {
		uuid := strings.ToLower(strings.TrimSpace(m.UUID))
		if uuid == "" || seen[uuid] {
			continue
		}
		seen[uuid] = true
		stats.Total++
		if m.RuntimeSec > 0 || m.ProcSec > 0 || m.ProductivityPct > 0 {
			stats.Active++
		}
		switch normalizeAiMacState(m.MacState) {
		case "2":
			stats.Working++
			stats.On++
		case "1":
			stats.Online++
			stats.On++
		default:
			stats.Offline++
		}
	}
	return stats
}

func looksLikeMachineOnCountQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	onKeys := []string{
		"sedang on", "sedang nyala", "mesin on", "mesin nyala", "yang on",
		"yang nyala", "power on", "masih on", "masih nyala", "online",
		"sedang menyala", "berapa yang on", "berapa yang nyala",
		"jumlah on", "jumlah nyala", "mesin hidup", "yang hidup",
		"macstate", "sedang jalan", "yang running", "sedang running",
	}
	hasOn := false
	for _, k := range onKeys {
		if strings.Contains(ql, k) {
			hasOn = true
			break
		}
	}
	if !hasOn {
		// "berapa mesin on" / "mesin yang on berapa"
		if strings.Contains(ql, "mesin") &&
			(strings.Contains(ql, " on") || strings.HasSuffix(ql, " on") ||
				strings.Contains(ql, "nyala") || strings.Contains(ql, "menyala")) {
			hasOn = true
		}
	}
	if !hasOn {
		return false
	}
	return strings.Contains(ql, "berapa") ||
		strings.Contains(ql, "jumlah") ||
		strings.Contains(ql, "total") ||
		strings.Contains(ql, "ada berapa") ||
		strings.Contains(ql, "mesin")
}

func looksLikeMachineCountQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	if looksLikeMachineOnCountQuestion(q) {
		return true
	}
	// Koreksi user seperti "bukan 54" / "harusnya 54" sering ikut setelah salah hitung.
	if strings.Contains(ql, "bukan ") && (strings.Contains(ql, "mesin") ||
		strings.Contains(ql, "54") || strings.Contains(ql, "total")) {
		return true
	}
	keys := []string{
		"berapa mesin", "berapa total mesin", "total mesin", "jumlah mesin",
		"ada berapa mesin", "mesin ada berapa", "total ada berapa",
		"berapa unit mesin", "count mesin", "banyak mesin",
		"mesin di dashboard", "mesin terlihat", "mesin yang terlihat",
	}
	for _, k := range keys {
		if strings.Contains(ql, k) {
			return true
		}
	}
	// "total ada berapa" / "ada berapa" dekat kata mesin di history — tangkap pola pendek.
	if (strings.Contains(ql, "berapa") || strings.Contains(ql, "jumlah") || strings.Contains(ql, "total")) &&
		strings.Contains(ql, "mesin") {
		return true
	}
	return false
}

func answerMachineCountQuestion(question string, req models.AiChatRequest) (string, bool) {
	if !looksLikeMachineCountQuestion(question) {
		return "", false
	}

	counts := countMachinesInRequest(req.Machines)
	loc := strings.TrimSpace(req.Location)
	if loc == "" {
		loc = "ALL"
	}
	shift := strings.TrimSpace(req.Shift)
	if shift == "" {
		shift = "-"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "ALL"
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		date = currentAiClock().Date
	}

	var b strings.Builder
	if looksLikeMachineOnCountQuestion(question) {
		b.WriteString(fmt.Sprintf(
			"Mesin yang sedang ON di filter dashboard: **%d** dari %d mesin.\n",
			counts.On,
			counts.Total,
		))
		b.WriteString(fmt.Sprintf(
			"Rincian status: %d working (proses) · %d online/idle · %d offline.\n\n",
			counts.Working,
			counts.Online,
			counts.Offline,
		))
		b.WriteString("ON = MacState online (1) atau working (2) dari data dashboard saat ini.\n")
	} else {
		b.WriteString(fmt.Sprintf(
			"Total mesin di filter dashboard saat ini: **%d**.\n",
			counts.Total,
		))
		b.WriteString(fmt.Sprintf(
			"Sedang ON: **%d** (working %d · online/idle %d) · Offline: %d.\n",
			counts.On,
			counts.Working,
			counts.Online,
			counts.Offline,
		))
		b.WriteString(fmt.Sprintf(
			"Punya data Power On/Process/produktivitas terukur: %d; belum ada jejak kerja: %d.\n\n",
			counts.Active,
			maxInt(0, counts.Total-counts.Active),
		))
	}
	b.WriteString(fmt.Sprintf(
		"Filter aktif: tanggal %s · area/line %s · shift %s · status %s.\n",
		date,
		loc,
		shift,
		status,
	))
	b.WriteString("Angka mengikuti mesin yang tampil di tabel dashboard (filter saat ini).")
	return strings.TrimSpace(b.String()), true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func answerClockQuestion(question, dashboardDate string) (string, bool) {
	if !looksLikeClockQuestion(question) {
		return "", false
	}
	now := currentAiClock()
	dash := strings.TrimSpace(dashboardDate)
	if dash == "" {
		dash = now.Date
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"Sekarang %s, tanggal %s, pukul %s (%s).",
		now.DayNameId,
		now.Date,
		now.Time,
		now.Timezone,
	))
	if dash != now.Date {
		b.WriteString(fmt.Sprintf(
			"\n\nCatatan: filter dashboard sedang melihat tanggal %s (bukan hari ini).",
			dash,
		))
	} else {
		b.WriteString("\n\nFilter dashboard juga memakai tanggal hari ini.")
	}
	return b.String(), true
}

var topNRegex = regexp.MustCompile(`(?i)\b(\d{1,2})\b`)

func parseTopN(q string, fallback int) int {
	m := topNRegex.FindStringSubmatch(q)
	if len(m) >= 2 {
		n, err := strconv.Atoi(m[1])
		if err == nil && n > 0 {
			if n > 20 {
				return 20
			}
			return n
		}
	}
	if fallback <= 0 {
		return 5
	}
	return fallback
}

func machineProductivityPct(m models.AiInsightMachineInput) float64 {
	return resolveAiUtilPct(m)
}

func activeMachinesForRank(machines []models.AiInsightMachineInput) []models.AiInsightMachineInput {
	out := make([]models.AiInsightMachineInput, 0, len(machines))
	for _, m := range machines {
		if m.RuntimeSec <= 0 && m.ProcSec <= 0 && m.ProductivityPct <= 0 {
			continue
		}
		out = append(out, m)
	}
	return out
}

func looksLikeTopProductiveQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	rankHint := strings.Contains(ql, "produktif") ||
		strings.Contains(ql, "produktivitas") ||
		strings.Contains(ql, "ranking") ||
		strings.Contains(ql, "rank ") ||
		strings.Contains(ql, "peringkat") ||
		strings.Contains(ql, "utilitas tertinggi") ||
		strings.Contains(ql, "util tertinggi")
	if !rankHint {
		return false
	}
	if strings.Contains(ql, "kurang") ||
		strings.Contains(ql, "terendah") ||
		strings.Contains(ql, "paling rendah") ||
		strings.Contains(ql, "buruk") ||
		strings.Contains(ql, "perhatian") {
		return false
	}
	return strings.Contains(ql, "paling") ||
		strings.Contains(ql, "tertinggi") ||
		strings.Contains(ql, "top") ||
		strings.Contains(ql, "terbaik") ||
		strings.Contains(ql, "tampilkan") ||
		strings.Contains(ql, "sebutkan") ||
		strings.Contains(ql, "ranking") ||
		strings.Contains(ql, "peringkat")
}

func looksLikeLowProductiveQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	return strings.Contains(ql, "kurang produktif") ||
		strings.Contains(ql, "paling rendah") ||
		strings.Contains(ql, "produktivitas terendah") ||
		strings.Contains(ql, "utilitas terendah") ||
		strings.Contains(ql, "mesin terburuk") ||
		(strings.Contains(ql, "produktif") && (strings.Contains(ql, "rendah") || strings.Contains(ql, "buruk")))
}

func formatMachineRankReply(title string, machines []models.AiInsightMachineInput, limit int, ascending bool) string {
	list := activeMachinesForRank(machines)
	if len(list) == 0 {
		return "Belum ada mesin aktif pada filter saat ini untuk di-ranking. Longgarkan filter atau refresh dashboard, lalu tanya lagi."
	}

	sort.SliceStable(list, func(i, j int) bool {
		pi := machineProductivityPct(list[i])
		pj := machineProductivityPct(list[j])
		if ascending {
			if pi == pj {
				return strings.ToLower(list[i].MachineName) < strings.ToLower(list[j].MachineName)
			}
			return pi < pj
		}
		if pi == pj {
			return strings.ToLower(list[i].MachineName) < strings.ToLower(list[j].MachineName)
		}
		return pi > pj
	})

	if limit <= 0 {
		limit = 5
	}
	if limit > len(list) {
		limit = len(list)
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	for i := 0; i < limit; i++ {
		m := list[i]
		name := strings.TrimSpace(m.MachineName)
		if name == "" {
			name = "Mesin"
		}
		loc := strings.TrimSpace(m.Location)
		line := fmt.Sprintf("%d. %s", i+1, name)
		if loc != "" {
			line += " — " + loc
		}
		line += fmt.Sprintf(" · produktivitas %.2f%%", machineProductivityPct(m))
		op := strings.TrimSpace(m.OperatorName)
		if op != "" {
			line += " · operator " + op
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\nAcuan: produktivitas mesin pada filter dashboard saat ini (bukan daftar prioritas Insight).")
	return strings.TrimSpace(b.String())
}

func answerRankingQuestion(question string, machines []models.AiInsightMachineInput) (string, bool) {
	n := parseTopN(question, 5)
	if looksLikeTopProductiveQuestion(question) {
		return formatMachineRankReply(
			fmt.Sprintf("%d mesin paling produktif saat ini:", n),
			machines,
			n,
			false,
		), true
	}
	if looksLikeLowProductiveQuestion(question) {
		return formatMachineRankReply(
			fmt.Sprintf("%d mesin dengan produktivitas terendah saat ini:", n),
			machines,
			n,
			true,
		), true
	}
	return "", false
}

func looksLikeOpenChartQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	return strings.Contains(ql, "chart") ||
		strings.Contains(ql, "grafik") ||
		strings.Contains(ql, "diagram") ||
		(strings.Contains(ql, "tampilkan") && strings.Contains(ql, "produktivitas"))
}

func findMachineForChart(question string, machines []models.AiInsightMachineInput) (models.AiInsightMachineInput, bool) {
	ql := strings.ToLower(strings.TrimSpace(question))
	bestIdx := -1
	bestScore := 0

	for i, m := range machines {
		name := strings.ToLower(strings.TrimSpace(m.MachineName))
		loc := strings.ToLower(strings.TrimSpace(m.Location))
		uuid := strings.ToLower(strings.TrimSpace(m.UUID))
		score := 0
		if name != "" && strings.Contains(ql, name) {
			score += len(name)
		}
		if loc != "" && strings.Contains(ql, loc) {
			score += len(loc)
		}
		if uuid != "" && strings.Contains(ql, uuid) {
			score += 50
		}
		// Token match kasar untuk nama pendek.
		for _, part := range strings.Fields(name) {
			if len(part) >= 4 && strings.Contains(ql, part) {
				score += len(part)
			}
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	if bestIdx < 0 || bestScore < 4 {
		return models.AiInsightMachineInput{}, false
	}
	return machines[bestIdx], true
}

func looksLikeNextHourActionQuestion(q string) bool {
	ql := strings.ToLower(strings.TrimSpace(q))
	if ql == "" {
		return false
	}
	hasTime := strings.Contains(ql, "1 jam") ||
		strings.Contains(ql, "satu jam") ||
		strings.Contains(ql, "ke depan")
	hasAction := strings.Contains(ql, "harus dilakukan") ||
		strings.Contains(ql, "apa yang harus") ||
		strings.Contains(ql, "prioritas") ||
		strings.Contains(ql, "aksi") ||
		strings.Contains(ql, "intervensi") ||
		strings.Contains(ql, "keputusan") ||
		strings.Contains(ql, "dilakukan")
	return (hasTime && hasAction) ||
		strings.Contains(ql, "prioritas 1 jam") ||
		strings.Contains(ql, "aksi 1 jam") ||
		(strings.Contains(ql, "1 jam") && strings.Contains(ql, "berdasarkan data"))
}

func insightActionTitle(item models.AiInsightItem) string {
	title := strings.TrimSpace(item.Location)
	if title == "" {
		title = strings.TrimSpace(item.MachineName)
	}
	if title == "" {
		title = strings.TrimSpace(item.Code)
	}
	if title == "" {
		title = "Item"
	}
	return title
}

func insightActionStep(item models.AiInsightItem) string {
	step := strings.TrimSpace(item.Improvement)
	if step == "" {
		step = strings.TrimSpace(item.Decision)
	}
	if step == "" {
		step = strings.TrimSpace(item.Signal)
	}
	if step == "" {
		step = "Follow-up utilisasi mesin dan lengkapi keterangan loss."
	}
	return step
}

func buildNextHourActionPlan(req models.AiChatRequest) string {
	var b strings.Builder
	b.WriteString("RENCANA AKSI 1 JAM KE DEPAN\n")
	b.WriteString("(berdasarkan AI Insight + data mesin aktif)\n\n")

	if req.Insight != nil && strings.TrimSpace(req.Insight.Summary) != "" {
		b.WriteString("Ringkas situasi:\n")
		b.WriteString(strings.TrimSpace(req.Insight.Summary))
		b.WriteString("\n\n")
	}

	items := []models.AiInsightItem{}
	if req.Insight != nil {
		items = req.Insight.Items
	}

	if len(items) == 0 {
		active := 0
		bad := 0
		for _, m := range req.Machines {
			if m.RuntimeSec <= 0 && m.ProcSec <= 0 && m.ProductivityPct <= 0 {
				continue
			}
			active++
			if strings.EqualFold(strings.TrimSpace(m.Status), "BAD") {
				bad++
			}
		}
		if active == 0 {
			return "Belum ada mesin aktif pada filter saat ini. Longgarkan filter / refresh dashboard, lalu tanya lagi."
		}
		b.WriteString(fmt.Sprintf(
			"Tidak ada prioritas Insight mendesak saat ini (%d mesin aktif, %d BAD).\n",
			active,
			bad,
		))
		b.WriteString("Lakukan:\n")
		b.WriteString("1. Pertahankan supervise rutin di line dengan utilitas terendah.\n")
		b.WriteString("2. Pastikan operator mengisi keterangan loss bila idle.\n")
		b.WriteString("3. Klik Refresh Insight 5–10 menit lagi untuk cek sinyal baru.\n")
		return strings.TrimSpace(b.String())
	}

	b.WriteString("Lakukan sekarang (urut prioritas):\n")
	limit := 7
	if len(items) < limit {
		limit = len(items)
	}
	for i := 0; i < limit; i++ {
		item := items[i]
		title := insightActionTitle(item)
		step := insightActionStep(item)
		b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, strings.ToUpper(strings.TrimSpace(item.Priority)), title))
		b.WriteString("   Lakukan: " + step + "\n")
		if item.UtilPct > 0 {
			b.WriteString(fmt.Sprintf("   Data: utilisasi %.1f%%\n", item.UtilPct))
		}
		if ev := strings.TrimSpace(item.Evidence); ev != "" {
			b.WriteString("   Bukti: " + ev + "\n")
		}
		if op := strings.TrimSpace(item.OperatorName); op != "" {
			b.WriteString("   Operator: " + op + "\n")
		}
		if note := strings.TrimSpace(item.NoteSummary); note != "" {
			b.WriteString("   Catatan: " + note + "\n")
		}
		if dec := strings.TrimSpace(item.Decision); dec != "" && !strings.EqualFold(dec, step) {
			b.WriteString("   Keputusan: " + dec + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("Fokus 60 menit ke depan: kerjakan nomor 1–3 dulu, baru lanjut sisanya.")
	return strings.TrimSpace(b.String())
}

func answerNextHourActionQuestion(question string, req models.AiChatRequest) (string, bool) {
	if !looksLikeNextHourActionQuestion(question) {
		return "", false
	}
	return buildNextHourActionPlan(req), true
}

func answerOpenChartQuestion(question string, machines []models.AiInsightMachineInput) (string, []models.AiChatAction, bool) {
	if !looksLikeOpenChartQuestion(question) {
		return "", nil, false
	}

	m, ok := findMachineForChart(question, machines)
	if !ok {
		// Kalau minta chart tanpa nama jelas: tawarkan top 1 produktif.
		list := activeMachinesForRank(machines)
		if len(list) == 0 {
			return "Belum ada mesin aktif untuk dibuka chart-nya. Longgarkan filter lalu coba lagi.", nil, true
		}
		sort.SliceStable(list, func(i, j int) bool {
			return machineProductivityPct(list[i]) > machineProductivityPct(list[j])
		})
		m = list[0]
	}

	name := strings.TrimSpace(m.MachineName)
	if name == "" {
		name = "Mesin"
	}
	loc := strings.TrimSpace(m.Location)
	title := name
	if loc != "" {
		title = name + " — " + loc
	}

	reply := fmt.Sprintf(
		"Siap. Saya arahkan ke chart produktivitas mesin:\n\n%s\nUtilitas/produktivitas: %.2f%%\n\nKlik tombol di bawah untuk membuka chart di dashboard.",
		title,
		machineProductivityPct(m),
	)
	actions := []models.AiChatAction{
		{
			Type:         "open_chart",
			Value:        strings.TrimSpace(m.UUID),
			Label:        "Buka chart: " + name,
			NeedsConfirm: true,
		},
	}
	return reply, normalizeAiChatActions(actions), true
}

// BuildAiChatContextJSON menyiapkan konteks aman untuk LLM (tanpa output/tiket).
func BuildAiChatContextJSON(req models.AiChatRequest) (string, error) {
	type machineView struct {
		UUID            string  `json:"uuid"`
		MachineName     string  `json:"machineName"`
		Location        string  `json:"location"`
		MacState        string  `json:"macState"`
		Status          string  `json:"status"`
		UtilPct         float64 `json:"utilPct"`
		ProductivityPct float64 `json:"productivityPct"`
		RuntimeSec      int64   `json:"runtimeSec"`
		ProcSec         int64   `json:"procSec"`
		LossTimeSec     int64   `json:"lossTimeSec"`
		OperatorName    string  `json:"operatorName,omitempty"`
		OperatorNotes   string  `json:"operatorNotes,omitempty"`
	}

	type insightItemView struct {
		Priority     string `json:"priority"`
		Scope        string `json:"scope,omitempty"`
		Code         string `json:"code"`
		Location     string `json:"location,omitempty"`
		MachineName  string `json:"machineName,omitempty"`
		Signal       string `json:"signal"`
		Evidence     string `json:"evidence,omitempty"`
		Decision     string `json:"decision,omitempty"`
		Improvement  string `json:"improvement"`
		NoteSummary  string `json:"noteSummary,omitempty"`
		OperatorName string `json:"operatorName,omitempty"`
	}

	counts := countMachinesInRequest(req.Machines)
	machines := make([]machineView, 0, len(req.Machines))
	for _, m := range req.Machines {
		if strings.TrimSpace(m.UUID) == "" {
			continue
		}
		// Detail list prioritas mesin yang punya jejak kerja; total tetap di machineCountTotal.
		if m.RuntimeSec <= 0 && m.ProcSec <= 0 && m.ProductivityPct <= 0 {
			continue
		}
		pct := machineProductivityPct(m)
		machines = append(machines, machineView{
			UUID:            strings.TrimSpace(m.UUID),
			MachineName:     strings.TrimSpace(m.MachineName),
			Location:        strings.TrimSpace(m.Location),
			MacState:        strings.TrimSpace(m.MacState),
			Status:          strings.TrimSpace(m.Status),
			UtilPct:         pct,
			ProductivityPct: pct,
			RuntimeSec:      m.RuntimeSec,
			ProcSec:         m.ProcSec,
			LossTimeSec:     m.LossTimeSec,
			OperatorName:    strings.TrimSpace(m.OperatorName),
			OperatorNotes:   strings.TrimSpace(m.OperatorNotes),
		})
		if len(machines) >= 40 {
			break
		}
	}

	var insightSummary string
	insightItems := make([]insightItemView, 0)
	if req.Insight != nil {
		insightSummary = strings.TrimSpace(req.Insight.Summary)
		for _, item := range req.Insight.Items {
			insightItems = append(insightItems, insightItemView{
				Priority:     item.Priority,
				Scope:        item.Scope,
				Code:         item.Code,
				Location:     item.Location,
				MachineName:  item.MachineName,
				Signal:       item.Signal,
				Evidence:     item.Evidence,
				Decision:     item.Decision,
				Improvement:  item.Improvement,
				NoteSummary:  item.NoteSummary,
				OperatorName: item.OperatorName,
			})
			if len(insightItems) >= 12 {
				break
			}
		}
	}

	nowInfo := currentAiClock()
	dashboardDate := strings.TrimSpace(req.Date)
	if dashboardDate == "" {
		dashboardDate = nowInfo.Date
	}

	payload := map[string]any{
		"page":               strings.TrimSpace(req.Page),
		"date":               dashboardDate,
		"now":                nowInfo,
		"shift":              strings.TrimSpace(req.Shift),
		"locationFilter":     strings.TrimSpace(req.Location),
		"statusFilter":       strings.TrimSpace(req.Status),
		"machineCount":       counts.Total, // alias total di filter dashboard
		"machineCountTotal":  counts.Total,
		"machineCountActive": counts.Active,
		"machineCountOn":     counts.On,
		"machineCountWorking": counts.Working,
		"machineCountOnline": counts.Online,
		"machineCountOffline": counts.Offline,
		"machinesInContext":  len(machines),
		"machines":           machines,
		"insightSummary":     insightSummary,
		"insightItems":       insightItems,
		"notes":              "Output/pcs dan tiket mekanik TIDAK disertakan. utilPct=productivityPct dashboard. Total=machineCountTotal. ON=machineCountOn (macState 1/2). machines[] bisa terpotong max 40. Field now = waktu server WIB.",
		"groundingRules": []string{
			"Total mesin di dashboard/filter = machineCountTotal",
			"Mesin sedang ON/nyala = machineCountOn (macState online+working)",
			"machineCountActive = mesin dengan Power On/Process/produktivitas > 0 (bukan sama dengan ON)",
			"Jangan bilang total = panjang machines[] saja",
			"IDLE_NO_NOTE = minta isi keterangan, bukan tebak penyebab",
			"Sebut nama mesin + util% + runtime/proc saat menganalisa",
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// CallOpenAIChat memanggil LLM (Ollama lokal atau OpenAI) dan parse jawaban terstruktur.
func CallOpenAIChat(ctx context.Context, req models.AiChatRequest) (models.AiChatResponse, error) {
	provider := AiChatProvider()
	if provider != "ollama" && provider != "openai" {
		return models.AiChatResponse{}, fmt.Errorf("AI_PROVIDER tidak didukung: %s (pakai ollama/openai)", provider)
	}

	if provider == "openai" {
		if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
			return models.AiChatResponse{}, fmt.Errorf("OPENAI_API_KEY belum di-set di .env")
		}
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
		return models.AiChatResponse{}, fmt.Errorf("pertanyaan wajib diisi")
	}

	// Pertanyaan kalender/jam: jawab langsung dari jam server (model kecil sering bilang "tidak ada data").
	if reply, ok := answerClockQuestion(question, req.Date); ok {
		return models.AiChatResponse{
			Status:  "ok",
			Reply:   reply,
			Actions: []models.AiChatAction{},
			Model:   "server-clock",
		}, nil
	}

	// Jumlah mesin: hitung dari payload dashboard, jangan andalkan LLM yang sering salah hitung.
	if reply, ok := answerMachineCountQuestion(question, req); ok {
		return models.AiChatResponse{
			Status:  "ok",
			Reply:   reply,
			Actions: []models.AiChatAction{},
			Model:   "server-count",
		}, nil
	}

	// Ranking produktif: hitung dari data mesin, jangan pakai fallback Insight HIGH.
	if reply, ok := answerRankingQuestion(question, req.Machines); ok {
		return models.AiChatResponse{
			Status:  "ok",
			Reply:   reply,
			Actions: []models.AiChatAction{},
			Model:   "server-rank",
		}, nil
	}

	// Buka chart mesin: arahkan ke modal chart dashboard.
	if reply, actions, ok := answerOpenChartQuestion(question, req.Machines); ok {
		return models.AiChatResponse{
			Status:  "ok",
			Reply:   reply,
			Actions: actions,
			Model:   "server-chart",
		}, nil
	}

	// Aksi 1 jam ke depan: pakai Insight (solusi konkret), jangan andalkan LLM yang sering generik.
	if reply, ok := answerNextHourActionQuestion(question, req); ok {
		return models.AiChatResponse{
			Status:  "ok",
			Reply:   reply,
			Actions: []models.AiChatAction{},
			Model:   "server-insight-plan",
		}, nil
	}

	contextJSON, err := BuildAiChatContextJSON(req)
	if err != nil {
		return models.AiChatResponse{}, err
	}

	messages := []openAIChatMessage{
		{Role: "system", Content: BuildAiChatSystemPrompt()},
		{Role: "system", Content: "Konteks dashboard saat ini (JSON):\n" + contextJSON},
	}

	// History terbatas (hindari token besar).
	histCount := 0
	for _, h := range req.History {
		role := strings.ToLower(strings.TrimSpace(h.Role))
		content := strings.TrimSpace(h.Content)
		if content == "" {
			continue
		}
		if role != "user" && role != "assistant" {
			continue
		}
		messages = append(messages, openAIChatMessage{Role: role, Content: content})
		histCount++
		if histCount >= 8 {
			break
		}
	}

	// Dorong JSON sebelum pertanyaan user (lebih efektif di Ollama).
	if provider == "ollama" {
		messages = append(messages, openAIChatMessage{
			Role: "system",
			Content: "PENTING: balas HANYA satu objek JSON valid. Field wajib: reply (string tidak kosong), actions (array). " +
				`Jawaban HARUS sesuai pertanyaan user. Jika ranking produktif, pakai productivityPct/utilPct dari machines — jangan salin insightItems. ` +
				`Contoh: {"reply":"5 mesin paling produktif: 1) Mesin A 80% ...","actions":[]}`,
		})
	}

	messages = append(messages, openAIChatMessage{Role: "user", Content: question})

	model := AiChatModel()
	body := openAIChatRequest{
		Model:       model,
		Temperature: 0.2,
		Messages:    messages,
	}
	// OpenAI: json_object lebih konsisten. Ollama: andalkan instruksi prompt.
	if provider == "openai" {
		body.ResponseFormat = map[string]string{"type": "json_object"}
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		return models.AiChatResponse{}, err
	}

	endpoint := aiChatCompletionsURL()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawBody))
	if err != nil {
		return models.AiChatResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if provider == "openai" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(os.Getenv("OPENAI_API_KEY")))
	} else {
		// Beberapa setup Ollama menerima bearer opsional.
		httpReq.Header.Set("Authorization", "Bearer ollama")
	}

	timeout := 120 * time.Second
	if provider == "openai" {
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	httpRes, err := client.Do(httpReq)
	if err != nil {
		if provider == "ollama" {
			return models.AiChatResponse{}, fmt.Errorf(
				"gagal hubungi Ollama di %s — pastikan Ollama sudah dijalankan dan model sudah di-pull (ollama pull %s): %v",
				ollamaBaseURL(),
				model,
				err,
			)
		}
		return models.AiChatResponse{}, fmt.Errorf("gagal hubungi OpenAI: %w", err)
	}
	defer httpRes.Body.Close()

	respRaw, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return models.AiChatResponse{}, err
	}

	var apiResp openAIChatAPIResponse
	if err := json.Unmarshal(respRaw, &apiResp); err != nil {
		return models.AiChatResponse{}, fmt.Errorf("respons %s tidak valid: %w", provider, err)
	}
	if apiResp.Error != nil && strings.TrimSpace(apiResp.Error.Message) != "" {
		return models.AiChatResponse{}, fmt.Errorf("%s: %s", strings.ToUpper(provider), apiResp.Error.Message)
	}
	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		return models.AiChatResponse{}, fmt.Errorf(
			"%s HTTP %d: %s",
			strings.ToUpper(provider),
			httpRes.StatusCode,
			strings.TrimSpace(string(respRaw)),
		)
	}
	if len(apiResp.Choices) == 0 {
		return models.AiChatResponse{}, fmt.Errorf("%s tidak mengembalikan jawaban", provider)
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	parsed, err := parseOpenAIStructuredReply(content)

	actions := []models.AiChatAction{}
	reply := ""
	if err == nil {
		actions = normalizeAiChatActions(parsed.Actions)
		reply = strings.TrimSpace(parsed.Reply)
	} else if content != "" && !strings.HasPrefix(content, "{") {
		// Teks biasa dari Ollama.
		reply = content
	}

	// Jika reply kosong, susun sesuai jenis pertanyaan (jangan asal dump Insight).
	if reply == "" {
		reply = BuildCopilotFallbackReply(req)
	}
	if reply == "" && content != "" {
		reply = content
	}
	if reply == "" {
		reply = "Data dashboard yang dikirim masih kosong atau belum cukup untuk analisa 1 jam ke depan. Refresh AI Insight / pastikan filter menampilkan mesin aktif, lalu tanya lagi."
	}

	return models.AiChatResponse{
		Status:  "ok",
		Reply:   reply,
		Actions: actions,
		Model:   model + " @" + provider,
	}, nil
}

func parseOpenAIStructuredReply(content string) (openAIStructuredReply, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return openAIStructuredReply{}, fmt.Errorf("konten kosong")
	}

	tryParse := func(raw string) (openAIStructuredReply, error) {
		var parsed openAIStructuredReply
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return openAIStructuredReply{}, err
		}
		if strings.TrimSpace(parsed.Reply) == "" {
			// Ollama kadang pakai nama field lain.
			var generic map[string]any
			if err := json.Unmarshal([]byte(raw), &generic); err == nil {
				for _, key := range []string{"reply", "answer", "message", "response", "text", "jawaban"} {
					if v, ok := generic[key]; ok {
						if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
							parsed.Reply = strings.TrimSpace(s)
							break
						}
					}
				}
			}
		}
		return parsed, nil
	}

	if parsed, err := tryParse(content); err == nil {
		return parsed, nil
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if parsed, err := tryParse(content[start : end+1]); err == nil {
			return parsed, nil
		}
	}
	return openAIStructuredReply{}, fmt.Errorf("bukan JSON terstruktur")
}

// BuildCopilotFallbackReply jawaban cadangan dari insight/mesin bila LLM gagal isi reply.
func BuildCopilotFallbackReply(req models.AiChatRequest) string {
	if reply, ok := answerMachineCountQuestion(req.Question, req); ok {
		return reply
	}
	if reply, ok := answerRankingQuestion(req.Question, req.Machines); ok {
		return reply
	}
	if reply, ok := answerClockQuestion(req.Question, req.Date); ok {
		return reply
	}

	q := strings.ToLower(strings.TrimSpace(req.Question))
	wantsPriority := strings.Contains(q, "1 jam") ||
		strings.Contains(q, "prioritas") ||
		strings.Contains(q, "intervensi") ||
		strings.Contains(q, "keputusan") ||
		strings.Contains(q, "aksi") ||
		q == ""

	if wantsPriority {
		return buildNextHourActionPlan(req)
	}

	active := 0
	bad := 0
	for _, m := range req.Machines {
		if m.RuntimeSec <= 0 && m.ProcSec <= 0 {
			continue
		}
		active++
		if strings.EqualFold(strings.TrimSpace(m.Status), "BAD") {
			bad++
		}
	}
	if active == 0 {
		return "Belum ada mesin aktif pada filter saat ini (Power On & Process = 0). Longgarkan filter area/shift/status atau refresh dashboard, lalu tanya lagi."
	}
	return fmt.Sprintf(
		"Ada %d mesin aktif dipantau (%d berstatus BAD). AI Insight belum terkirim lengkap ke Copilot — klik Refresh pada panel AI Insight, lalu ulangi pertanyaan \"Aksi 1 jam ke depan\".",
		active,
		bad,
	)
}

func normalizeAiChatActions(actions []models.AiChatAction) []models.AiChatAction {
	allowed := map[string]bool{
		"set_location":    true,
		"set_shift":       true,
		"set_status":      true,
		"set_date":        true,
		"navigate":        true,
		"refresh_insight": true,
		"open_chart":      true,
	}

	out := make([]models.AiChatAction, 0, len(actions))
	for _, a := range actions {
		t := strings.ToLower(strings.TrimSpace(a.Type))
		v := strings.TrimSpace(a.Value)
		if !allowed[t] || v == "" {
			continue
		}
		label := strings.TrimSpace(a.Label)
		if label == "" {
			label = defaultAiActionLabel(t, v)
		}
		out = append(out, models.AiChatAction{
			Type:         t,
			Value:        v,
			Label:        label,
			NeedsConfirm: true, // selalu konfirmasi di V1
		})
		if len(out) >= 5 {
			break
		}
	}
	return out
}

func defaultAiActionLabel(actionType, value string) string {
	switch actionType {
	case "set_location":
		return "Terapkan area/line: " + value
	case "set_shift":
		return "Terapkan shift: " + value
	case "set_status":
		return "Terapkan status: " + value
	case "set_date":
		return "Terapkan tanggal: " + value
	case "navigate":
		return "Buka menu: " + value
	case "refresh_insight":
		return "Refresh AI Insight"
	case "open_chart":
		return "Buka chart mesin"
	default:
		return value
	}
}
