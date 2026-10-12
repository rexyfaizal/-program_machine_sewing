package utils

import (
	"fmt"
	"sort"
	"strings"

	"backend_machine/models"
)

const (
	AiPriorityHigh   = "HIGH"
	AiPriorityMedium = "MEDIUM"
	AiPriorityLow    = "LOW"

	AiLowUtilPct           = 55.0
	AiMinRuntimeForIdleSec = 20 * 60
	AiMinLossHighSec       = 5 * 60
	AiBrokenWindowDays     = 7
	AiFrequentBrokenMin    = 3 // ≥3 event MACHINE_BROKEN dalam 7 hari
)

type aiMachineContext struct {
	input         models.AiInsightMachineInput
	session       *models.AiInsightActiveSession
	loss          *models.AiInsightActiveLoss
	notes         []models.AiInsightRecentNote
	utilPct       float64
	macState      string
	hasLogin      bool
	brokenCount7d int
	finding       string // idle | loss_broken | frequent_broken | loss_hanca | loss_other | offline | no_login | bad | incomplete
}

type lineBucket struct {
	line     string
	findings map[string][]aiMachineContext
}

// BuildAiInsightItems menjalankan analisa keputusan (tanpa output, tanpa tiket).
func BuildAiInsightItems(
	machines []models.AiInsightMachineInput,
	sessions []models.AiInsightActiveSession,
	losses []models.AiInsightActiveLoss,
	notes []models.AiInsightRecentNote,
	brokenFreq []models.AiInsightBrokenFrequency,
	locationFilter string,
) []models.AiInsightItem {
	sessionByUUID := map[string]*models.AiInsightActiveSession{}
	for i := range sessions {
		key := normalizeAiUUID(sessions[i].UUID)
		if key == "" {
			continue
		}
		if _, exists := sessionByUUID[key]; !exists {
			sessionByUUID[key] = &sessions[i]
		}
	}

	lossByUUID := map[string]*models.AiInsightActiveLoss{}
	for i := range losses {
		key := normalizeAiUUID(losses[i].UUID)
		if key == "" {
			continue
		}
		if _, exists := lossByUUID[key]; !exists {
			lossByUUID[key] = &losses[i]
		}
	}

	notesByUUID := map[string][]models.AiInsightRecentNote{}
	for i := range notes {
		key := normalizeAiUUID(notes[i].UUID)
		if key == "" {
			continue
		}
		notesByUUID[key] = append(notesByUUID[key], notes[i])
	}

	brokenByUUID := map[string]int{}
	for _, row := range brokenFreq {
		key := normalizeAiUUID(row.UUID)
		if key == "" || row.Count <= 0 {
			continue
		}
		brokenByUUID[key] = row.Count
	}

	locationFilter = strings.ToUpper(strings.TrimSpace(locationFilter))
	lineMap := map[string]*lineBucket{}
	machineItems := make([]models.AiInsightItem, 0)

	for _, machine := range machines {
		key := normalizeAiUUID(machine.UUID)
		if key == "" {
			continue
		}
		brokenCount := brokenByUUID[key]
		// Tetap analisa mesin tanpa jejak hari ini jika sering rusak (riwayat 7 hari).
		if machine.RuntimeSec <= 0 && machine.ProcSec <= 0 && brokenCount < AiFrequentBrokenMin {
			continue
		}
		if locationFilter != "" && locationFilter != "ALL" {
			loc := strings.ToUpper(strings.TrimSpace(machine.Location))
			if loc == "" || !strings.Contains(loc, locationFilter) {
				continue
			}
		}

		ctx := aiMachineContext{
			input:         machine,
			macState:      normalizeAiMacState(machine.MacState),
			utilPct:       resolveAiUtilPct(machine),
			notes:         notesByUUID[key],
			brokenCount7d: brokenCount,
		}
		if s, ok := sessionByUUID[key]; ok {
			ctx.session = s
			ctx.hasLogin = true
		}
		if strings.TrimSpace(machine.OperatorName) != "" {
			ctx.hasLogin = true
		}
		if l, ok := lossByUUID[key]; ok {
			ctx.loss = l
		}

		item, ok := evaluateAiMachineDecision(ctx)
		if !ok {
			continue
		}
		machineItems = append(machineItems, item)

		// Simpan konteks finding untuk agregasi line.
		ctx.finding = mapCodeToFinding(item.Code, item.LossReasonCode, item.LossReasonLabel)
		lineKey := normalizeAiLineKey(machine.Location)
		if lineKey == "" {
			continue
		}
		bucket, exists := lineMap[lineKey]
		if !exists {
			bucket = &lineBucket{line: lineKey, findings: map[string][]aiMachineContext{}}
			lineMap[lineKey] = bucket
		}
		bucket.findings[ctx.finding] = append(bucket.findings[ctx.finding], ctx)
	}

	items := make([]models.AiInsightItem, 0)

	// Keputusan level line (lebih bernilai untuk SPV).
	lineItems := buildLineDecisions(lineMap)
	items = append(items, lineItems...)

	// Detail mesin: ambil yang paling kritis, hindari spam.
	sort.SliceStable(machineItems, func(i, j int) bool {
		if machineItems[i].RiskScore != machineItems[j].RiskScore {
			return machineItems[i].RiskScore > machineItems[j].RiskScore
		}
		return machineItems[i].Code < machineItems[j].Code
	})

	coveredLines := map[string]bool{}
	for _, li := range lineItems {
		coveredLines[normalizeAiLineKey(li.Location)] = true
	}

	detailLimit := 8
	added := 0
	for _, item := range machineItems {
		if added >= detailLimit {
			break
		}
		// Jika line sudah punya keputusan, hanya tampilkan mesin HIGH unik (loss/offline).
		line := normalizeAiLineKey(item.Location)
		if coveredLines[line] && item.Priority != AiPriorityHigh {
			continue
		}
		if coveredLines[line] && item.Code == "IDLE_NO_NOTE" {
			continue
		}
		items = append(items, item)
		added++
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RiskScore != items[j].RiskScore {
			return items[i].RiskScore > items[j].RiskScore
		}
		// Line decision dulu pada skor sama.
		if items[i].Scope != items[j].Scope {
			return items[i].Scope == "LINE"
		}
		return items[i].Code < items[j].Code
	})

	if len(items) > 15 {
		items = items[:15]
	}
	return items
}

func evaluateAiMachineDecision(ctx aiMachineContext) (models.AiInsightItem, bool) {
	name := strings.TrimSpace(ctx.input.MachineName)
	if name == "" {
		name = strings.TrimSpace(ctx.input.UUID)
	}
	loc := strings.TrimSpace(ctx.input.Location)
	uuid := strings.TrimSpace(ctx.input.UUID)

	base := models.AiInsightItem{
		Scope:         "MACHINE",
		UUID:          uuid,
		MachineName:   name,
		Location:      loc,
		UtilPct:       ctx.utilPct,
		BrokenCount7d: ctx.brokenCount7d,
	}
	if ctx.session != nil {
		base.OperatorName = strings.TrimSpace(ctx.session.OperatorName)
	}
	if base.OperatorName == "" {
		base.OperatorName = strings.TrimSpace(ctx.input.OperatorName)
	}
	if ctx.loss != nil {
		base.LossReasonCode = strings.ToUpper(strings.TrimSpace(ctx.loss.ReasonCode))
		base.LossReasonLabel = strings.TrimSpace(ctx.loss.ReasonLabel)
		if base.OperatorName == "" {
			base.OperatorName = strings.TrimSpace(ctx.loss.OperatorName)
		}
	}

	noteSummary := buildNoteSummary(ctx)
	base.NoteSummary = noteSummary

	opLabel := base.OperatorName
	if opLabel == "" {
		opLabel = "operator"
	}

	// 1) Offline
	if ctx.macState == "0" {
		item := base
		item.Priority = AiPriorityHigh
		item.Code = "MACHINE_OFFLINE"
		item.Signal = fmt.Sprintf("%s offline saat masih ada jejak kerja hari ini", name)
		item.Evidence = fmt.Sprintf(
			"MacState=0 · Power On %s · Process %s",
			formatAiDuration(ctx.input.RuntimeSec),
			formatAiDuration(ctx.input.ProcSec),
		)
		item.Prediction1h = "Tanpa recovery, mesin ini hampir pasti tidak menyumbang process time 1 jam ke depan"
		item.Decision = "Prioritaskan cek fisik mesin (power/jaringan) sebelum supervise utilitas"
		item.Improvement = fmt.Sprintf("Teknisi/line leader cek %s di %s sekarang", name, fallbackText(loc, "area mesin"))
		item.RiskScore = 88
		return item, true
	}

	// 2) Mesin ON / power on / process jalan, tapi operator belum login
	if !ctx.hasLogin && isNoLoginRunning(ctx) {
		item := base
		item.Code = "NO_LOGIN_RUNNING"
		running := ctx.input.ProcSec > 0 || ctx.macState == "2"
		item.Priority = AiPriorityMedium
		item.RiskScore = 74
		activity := "menyala/online"
		if running {
			item.Priority = AiPriorityHigh
			item.RiskScore = 88
			activity = "berjalan (process aktif)"
		}
		item.Signal = fmt.Sprintf("%s %s tanpa sesi operator login", name, activity)
		item.Evidence = joinAiParts(
			fmt.Sprintf("MacState=%s · tidak ada login aktif", fallbackText(ctx.macState, "-")),
			fmt.Sprintf(
				"Power On %s · Process %s · Utilisasi dashboard %.1f%%",
				formatAiDuration(ctx.input.RuntimeSec),
				formatAiDuration(ctx.input.ProcSec),
				ctx.utilPct,
			),
			"Anomali: mesin ON/jalan tanpa operator terikat",
		)
		if running {
			item.Prediction1h = fmt.Sprintf(
				"Process %s di %s tidak terikat operator — risiko output tidak accountable 1 jam ke depan",
				formatAiDuration(ctx.input.ProcSec),
				name,
			)
			item.Decision = fmt.Sprintf(
				"Segera login operator atau hentikan pemakaian %s — mesin jalan tanpa login",
				name,
			)
			item.Improvement = fmt.Sprintf(
				"1. Line leader cek siapa yang pakai %s di %s\n2. Paksa login operator aktif\n3. Jika tidak dipakai: matikan mesin",
				name,
				fallbackText(loc, "area mesin"),
			)
		} else {
			item.Prediction1h = "Produktivitas tidak terikat operator; risiko nyala sia-sia 1 jam ke depan"
			item.Decision = "Rapikan disiplin login atau matikan mesin yang tidak dipakai"
			item.Improvement = fmt.Sprintf(
				"1. Pastikan %s di-login operator\n2. Atau matikan jika tidak dipakai",
				name,
			)
		}
		return item, true
	}

	// 3) Loss aktif + note
	if ctx.loss != nil {
		code := strings.ToUpper(strings.TrimSpace(ctx.loss.ReasonCode))
		label := strings.TrimSpace(ctx.loss.ReasonLabel)
		if label == "" {
			label = humanizeLossCode(code)
		}
		lossNote := strings.TrimSpace(ctx.loss.Note)
		item := base
		item.Code = "ACTIVE_LOSS"
		item.Signal = fmt.Sprintf(
			"%s masih dalam keterangan %s (%s)",
			opLabel,
			label,
			formatAiDuration(ctx.loss.DurationSec),
		)
		item.Evidence = joinAiParts(
			fmt.Sprintf(
				"Utilisasi dashboard %.1f%% · Power On %s · Process %s",
				ctx.utilPct,
				formatAiDuration(ctx.input.RuntimeSec),
				formatAiDuration(ctx.input.ProcSec),
			),
			fmt.Sprintf("Loss aktif %s sejak %s (%s)", label, fallbackText(ctx.loss.StartTime, "-"), formatAiDuration(ctx.loss.DurationSec)),
			noteSummary,
			lossNote,
		)

		switch classifyLoss(code, label) {
		case "broken":
			item.Priority = AiPriorityHigh
			item.Prediction1h = fmt.Sprintf(
				"Process %s di %s berisiko tetap terhambat sampai penanganan teknis selesai",
				formatAiDuration(ctx.input.ProcSec),
				name,
			)
			item.Decision = fmt.Sprintf(
				"Prioritas teknis: %s util %.1f%% masih loss %s",
				name,
				ctx.utilPct,
				label,
			)
			item.Improvement = fmt.Sprintf(
				"Pastikan penanganan %s (%s) berjalan; konfirmasi ke %s (util %.1f%%)",
				name,
				label,
				opLabel,
				ctx.utilPct,
			)
			item.RiskScore = 94
			if ctx.brokenCount7d >= AiFrequentBrokenMin {
				prev := buildFrequentBrokenPreventive(name, loc, ctx.brokenCount7d, true)
				item.Signal = fmt.Sprintf(
					"%s loss mesin rusak aktif — dan sering rusak %dx / %d hari",
					name,
					ctx.brokenCount7d,
					AiBrokenWindowDays,
				)
				item.Evidence = joinAiParts(
					item.Evidence,
					fmt.Sprintf("Riwayat MACHINE_BROKEN %dx dalam %d hari", ctx.brokenCount7d, AiBrokenWindowDays),
					"Rekomendasi: preventive sebelum shift berikutnya",
				)
				item.Decision = prev.decision
				item.Improvement = prev.improvement
				item.Prediction1h = prev.prediction
				item.RiskScore = 97
			}
		case "hanca":
			item.Priority = AiPriorityHigh
			item.Prediction1h = fmt.Sprintf(
				"Utilisasi %s (%.1f%%) berisiko stagnan selama %s belum selesai",
				name,
				ctx.utilPct,
				label,
			)
			item.Decision = fmt.Sprintf(
				"Intervensi supply sesuai reason %s — lebih mendesak daripada coaching kecepatan",
				label,
			)
			item.Improvement = fmt.Sprintf(
				"SPV %s pastikan %s untuk %s di %s tersedia sekarang (loss %s)",
				fallbackText(normalizeAiLineKey(loc), loc),
				label,
				opLabel,
				name,
				formatAiDuration(ctx.loss.DurationSec),
			)
			item.RiskScore = 92
		default:
			item.Priority = AiPriorityMedium
			item.Prediction1h = fmt.Sprintf(
				"Selama %s belum selesai, kontribusi process %s 1 jam ke depan terbatas",
				label,
				name,
			)
			item.Decision = fmt.Sprintf(
				"Monitor loss %s di %s (util %.1f%%); eskalasi jika melebihi batas wajar line",
				label,
				name,
				ctx.utilPct,
			)
			item.Improvement = fmt.Sprintf(
				"Follow-up %s di %s: selesaikan keterangan %s (berlangsung %s, util %.1f%%)",
				opLabel,
				name,
				label,
				formatAiDuration(ctx.loss.DurationSec),
				ctx.utilPct,
			)
			item.RiskScore = 74
			if ctx.loss.DurationSec >= AiMinLossHighSec {
				item.Priority = AiPriorityHigh
				item.RiskScore = 84
			}
		}
		return item, true
	}

	// 2b) Sering rusak (riwayat 7 hari) meski loss aktif sudah tidak ada
	if ctx.brokenCount7d >= AiFrequentBrokenMin {
		prev := buildFrequentBrokenPreventive(name, loc, ctx.brokenCount7d, false)
		item := base
		item.Priority = AiPriorityHigh
		item.Code = "FREQUENT_BROKEN"
		item.Signal = fmt.Sprintf(
			"%s sering rusak: %dx dalam %d hari terakhir",
			name,
			ctx.brokenCount7d,
			AiBrokenWindowDays,
		)
		item.Evidence = joinAiParts(
			fmt.Sprintf(
				"Riwayat MACHINE_BROKEN / Mesin Rusak: %dx (jendela %d hari)",
				ctx.brokenCount7d,
				AiBrokenWindowDays,
			),
			fmt.Sprintf(
				"Utilisasi dashboard %.1f%% · Power On %s · Process %s",
				ctx.utilPct,
				formatAiDuration(ctx.input.RuntimeSec),
				formatAiDuration(ctx.input.ProcSec),
			),
			"Preventive: cek teknis sebelum shift berikutnya",
			noteSummary,
		)
		item.Prediction1h = prev.prediction
		item.Decision = prev.decision
		item.Improvement = prev.improvement
		item.RiskScore = 90
		if ctx.brokenCount7d >= 5 {
			item.RiskScore = 95
		}
		return item, true
	}

	// 4) Ada note terbaru tapi util rendah → pakai note sebagai konteks
	if noteSummary != "" &&
		ctx.input.RuntimeSec >= AiMinRuntimeForIdleSec &&
		ctx.utilPct < AiLowUtilPct {
		item := base
		item.Priority = AiPriorityMedium
		item.Code = "NOTE_CONTEXT_LOW_UTIL"
		item.Signal = fmt.Sprintf(
			"%s di %s: utilisasi %.1f%% dengan keterangan: %s",
			opLabel,
			fallbackText(loc, name),
			ctx.utilPct,
			clipAiText(noteSummary, 70),
		)
		item.Evidence = joinAiParts(
			fmt.Sprintf(
				"Utilisasi dashboard %.1f%% · Power On %s · Process %s · Loss sistem %s",
				ctx.utilPct,
				formatAiDuration(ctx.input.RuntimeSec),
				formatAiDuration(ctx.input.ProcSec),
				formatAiDuration(ctx.input.LossTimeSec),
			),
			"Keterangan: "+noteSummary,
		)
		item.Prediction1h = fmt.Sprintf(
			"Selama penyebab di keterangan belum selesai, utilisasi %s (%.1f%%) berisiko tetap tertekan 1 jam ke depan",
			name,
			ctx.utilPct,
		)
		item.Decision = fmt.Sprintf(
			"Tindaklanjuti sesuai keterangan yang sudah ada (%s), bukan asumsi idle umum",
			clipAiText(noteSummary, 50),
		)
		item.Improvement = fmt.Sprintf(
			"SPV verifikasi keterangan %s di %s (util %.1f%%): %s",
			opLabel,
			fallbackText(normalizeAiLineKey(loc), name),
			ctx.utilPct,
			clipAiText(noteSummary, 60),
		)
		item.RiskScore = 76
		if ctx.utilPct < 35 {
			item.Priority = AiPriorityHigh
			item.RiskScore = 86
		}
		return item, true
	}

	// 5) Idle tanpa keterangan — hanya celah data, tanpa spekulasi penyebab
	if ctx.hasLogin &&
		ctx.input.RuntimeSec >= AiMinRuntimeForIdleSec &&
		ctx.utilPct > 0 &&
		ctx.utilPct < AiLowUtilPct &&
		noteSummary == "" {
		item := base
		item.Priority = AiPriorityHigh
		item.Code = "IDLE_NO_NOTE"
		copy := buildIdleNoNoteCopy(ctx, opLabel, name, loc)
		item.Signal = copy.signal
		item.Evidence = copy.evidence
		item.Prediction1h = copy.prediction
		item.Decision = copy.decision
		item.Improvement = copy.improvement
		item.RiskScore = copy.risk
		return item, true
	}

	// 6) Status BAD tanpa finding lain
	status := strings.ToUpper(strings.TrimSpace(ctx.input.Status))
	if status == "BAD" {
		item := base
		item.Priority = AiPriorityMedium
		item.Code = "STATUS_BAD"
		item.Signal = fmt.Sprintf(
			"%s di %s status BAD: utilisasi %.1f%% (process %s / power on %s)",
			name,
			fallbackText(loc, "-"),
			ctx.utilPct,
			formatAiDuration(ctx.input.ProcSec),
			formatAiDuration(ctx.input.RuntimeSec),
		)
		item.Evidence = joinAiParts(
			fmt.Sprintf(
				"Utilisasi dashboard %.1f%% · Power On %s · Process %s · Loss sistem %s · Status BAD",
				ctx.utilPct,
				formatAiDuration(ctx.input.RuntimeSec),
				formatAiDuration(ctx.input.ProcSec),
				formatAiDuration(ctx.input.LossTimeSec),
			),
			noteSummary,
		)
		item.Prediction1h = fmt.Sprintf(
			"Tanpa tinjauan, %s berisiko tetap di bawah standar (util %.1f%%) pada 1 jam ke depan",
			name,
			ctx.utilPct,
		)
		item.Decision = fmt.Sprintf(
			"Watchlist supervise: %s util %.1f%% — tinjau gap process vs power on sebelum eskalasi",
			name,
			ctx.utilPct,
		)
		item.Improvement = fmt.Sprintf(
			"SPV tinjau %s bersama %s (util %.1f%%, loss sistem %s)",
			name,
			opLabel,
			ctx.utilPct,
			formatAiDuration(ctx.input.LossTimeSec),
		)
		item.RiskScore = 68
		return item, true
	}

	// 7) Style/proses kosong
	if ctx.session != nil {
		style := strings.TrimSpace(ctx.session.StyleName)
		process := strings.TrimSpace(ctx.session.ProcessName)
		if style == "" || process == "" || style == "-" || process == "-" {
			item := base
			item.Priority = AiPriorityLow
			item.Code = "SESSION_DATA_INCOMPLETE"
			item.Signal = fmt.Sprintf("Sesi %s belum lengkap style/proses", opLabel)
			item.Evidence = fmt.Sprintf("Style=%s · Proses=%s", fallbackText(style, "-"), fallbackText(process, "-"))
			item.Prediction1h = "Monitoring sesi kurang akurat sampai data dilengkapi"
			item.Decision = "Lengkapi master data sesi sebelum evaluasi performa"
			item.Improvement = fmt.Sprintf("Koreksi style/proses untuk sesi %s", opLabel)
			item.RiskScore = 40
			return item, true
		}
	}

	return models.AiInsightItem{}, false
}

func buildLineDecisions(lineMap map[string]*lineBucket) []models.AiInsightItem {
	type scored struct {
		item  models.AiInsightItem
		count int
	}
	scoredLines := make([]scored, 0)

	for line, bucket := range lineMap {
		total := 0
		for _, list := range bucket.findings {
			total += len(list)
		}
		if total < 2 {
			continue
		}

		copy := buildLineDecisionCopy(line, bucket, total)
		if strings.TrimSpace(copy.decision) == "" {
			continue
		}

		item := models.AiInsightItem{
			Priority:     AiPriorityHigh,
			Scope:        "LINE",
			Code:         "LINE_DECISION",
			Location:     line,
			Signal:       copy.signal,
			Evidence:     copy.evidence,
			Prediction1h: copy.prediction,
			Decision:     copy.decision,
			Improvement:  copy.improvement,
			RiskScore:    copy.risk,
			NoteSummary:  copy.noteSummary,
		}
		if total < 3 && copy.risk < 85 {
			item.Priority = AiPriorityMedium
		}
		scoredLines = append(scoredLines, scored{item: item, count: total})
	}

	sort.SliceStable(scoredLines, func(i, j int) bool {
		if scoredLines[i].item.RiskScore != scoredLines[j].item.RiskScore {
			return scoredLines[i].item.RiskScore > scoredLines[j].item.RiskScore
		}
		return scoredLines[i].count > scoredLines[j].count
	})

	limit := 5
	if len(scoredLines) < limit {
		limit = len(scoredLines)
	}
	out := make([]models.AiInsightItem, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, scoredLines[i].item)
	}
	return out
}

func mapCodeToFinding(code, lossCode, lossLabel string) string {
	switch code {
	case "MACHINE_OFFLINE":
		return "offline"
	case "NO_LOGIN_RUNNING", "NO_OPERATOR":
		return "no_login"
	case "IDLE_NO_NOTE":
		return "idle"
	case "NOTE_CONTEXT_LOW_UTIL":
		return "note_low"
	case "FREQUENT_BROKEN":
		return "frequent_broken"
	case "ACTIVE_LOSS":
		switch classifyLoss(lossCode, lossLabel) {
		case "broken":
			return "loss_broken"
		case "hanca":
			return "loss_hanca"
		default:
			return "loss_other"
		}
	default:
		return "other"
	}
}

// isNoLoginRunning: operator belum login, tapi mesin ON / ada power-on / ada process.
func isNoLoginRunning(ctx aiMachineContext) bool {
	if ctx.hasLogin {
		return false
	}
	if ctx.macState == "1" || ctx.macState == "2" {
		return true
	}
	// Cadangan jika MacState kosong/tidak sinkron, tapi jejak kerja ada.
	if ctx.input.ProcSec > 0 {
		return true
	}
	if ctx.input.RuntimeSec >= AiMinRuntimeForIdleSec {
		return true
	}
	return false
}

func classifyLoss(code, label string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	l := strings.ToUpper(strings.TrimSpace(label))
	// Hanya reason yang jelas — hindari false positive dari kata generik.
	if c == "MACHINE_BROKEN" || strings.Contains(l, "MESIN RUSAK") {
		return "broken"
	}
	if c == "WAIT_HANCA" || strings.Contains(l, "HANCA") || strings.Contains(l, "TUNGGU BAHAN") {
		return "hanca"
	}
	return "other"
}

type idleNoNoteCopy struct {
	signal      string
	evidence    string
	prediction  string
	decision    string
	improvement string
	risk        int
}

// buildIdleNoNoteCopy menyusun teks berbeda per mesin dari metrik dashboard saja.
func buildIdleNoNoteCopy(ctx aiMachineContext, opLabel, name, loc string) idleNoNoteCopy {
	util := ctx.utilPct
	rt := formatAiDuration(ctx.input.RuntimeSec)
	proc := formatAiDuration(ctx.input.ProcSec)
	loss := formatAiDuration(ctx.input.LossTimeSec)
	line := fallbackText(normalizeAiLineKey(loc), loc)
	where := fallbackText(loc, name)

	lossRatio := 0.0
	if ctx.input.RuntimeSec > 0 {
		lossRatio = float64(ctx.input.LossTimeSec) * 100.0 / float64(ctx.input.RuntimeSec)
	}

	out := idleNoNoteCopy{
		evidence: fmt.Sprintf(
			"Utilisasi dashboard %.1f%% · Power On %s · Process %s · Loss sistem %s · Keterangan/loss aktif: tidak ada",
			util,
			rt,
			proc,
			loss,
		),
		risk: 87,
	}

	switch {
	case util < 20:
		out.signal = fmt.Sprintf(
			"%s di %s: utilisasi sangat rendah %.1f%% (process %s dari power on %s), belum ada keterangan",
			opLabel, where, util, proc, rt,
		)
		out.decision = fmt.Sprintf(
			"Celah data kritis di %s: utilisasi %.1f%% tanpa keterangan — jangan spekulasi penyebab sebelum note diisi",
			fallbackText(line, name),
			util,
		)
		out.improvement = fmt.Sprintf(
			"Minta %s isi keterangan loss di %s sekarang (util %.1f%%, process %s, loss sistem %s)",
			opLabel, fallbackText(line, name), util, proc, loss,
		)
		out.prediction = fmt.Sprintf(
			"Tanpa keterangan, gap process %s vs power on %s di %s tetap tidak bisa ditindaklanjuti 1 jam ke depan",
			proc, rt, name,
		)
		out.risk = 94
	case util < 35:
		out.signal = fmt.Sprintf(
			"%s (%s): process %s jauh di bawah power on %s — util %.1f%%, tanpa note/loss aktif",
			name, opLabel, proc, rt, util,
		)
		out.decision = fmt.Sprintf(
			"Celah data: util %.1f%% dengan loss sistem %s belum dijelaskan operator",
			util, loss,
		)
		out.improvement = fmt.Sprintf(
			"Minta %s lengkapi keterangan loss untuk %s (util %.1f%%) agar SPV bisa putuskan aksi",
			opLabel, name, util,
		)
		out.prediction = fmt.Sprintf(
			"Jika idle tersembunyi berlanjut, utilisasi %s (~%.1f%%) berisiko tetap tertekan 1 jam ke depan",
			name, util,
		)
		out.risk = 92
	case lossRatio >= 50:
		out.signal = fmt.Sprintf(
			"%s di %s: loss sistem %s (~%.0f%% dari power on) tanpa keterangan — util %.1f%%",
			opLabel, where, loss, lossRatio, util,
		)
		out.decision = fmt.Sprintf(
			"Celah data: loss sistem mendominasi power on di %s tanpa catatan operator",
			name,
		)
		out.improvement = fmt.Sprintf(
			"Minta %s isi keterangan untuk loss sistem %s pada %s (util dashboard %.1f%%)",
			opLabel, loss, name, util,
		)
		out.prediction = fmt.Sprintf(
			"Tanpa klarifikasi loss %s, keputusan utilisasi %s 1 jam ke depan tetap buta data",
			loss, name,
		)
		out.risk = 90
	default:
		out.signal = fmt.Sprintf(
			"%s di %s: utilisasi %.1f%% di bawah ambang (power on %s / process %s), belum ada keterangan",
			opLabel, where, util, rt, proc,
		)
		out.decision = fmt.Sprintf(
			"Celah data di %s: utilisasi %.1f%% tanpa note — lengkapi keterangan dulu sebelum menafsirkan penyebab",
			fallbackText(line, name),
			util,
		)
		out.improvement = fmt.Sprintf(
			"Minta %s isi keterangan loss sekarang di %s (util %.1f%%, process %s)",
			opLabel, fallbackText(line, name), util, proc,
		)
		out.prediction = fmt.Sprintf(
			"Tanpa note, sulit bedakan idle vs proses lambat di %s selama 1 jam ke depan",
			name,
		)
		out.risk = 87
	}

	return out
}

func sampleMachineUtilHints(list []aiMachineContext, limit int) []string {
	out := make([]string, 0, limit)
	for _, ctx := range list {
		name := strings.TrimSpace(ctx.input.MachineName)
		if name == "" {
			name = strings.TrimSpace(ctx.input.UUID)
		}
		op := strings.TrimSpace(ctx.input.OperatorName)
		if ctx.session != nil && strings.TrimSpace(ctx.session.OperatorName) != "" {
			op = strings.TrimSpace(ctx.session.OperatorName)
		}
		hint := fmt.Sprintf("%s %.1f%%", name, ctx.utilPct)
		if op != "" {
			hint = fmt.Sprintf("%s / %s %.1f%%", name, op, ctx.utilPct)
		}
		out = append(out, hint)
		if len(out) >= limit {
			break
		}
	}
	return out
}

type lineDecisionCopy struct {
	signal      string
	evidence    string
	prediction  string
	decision    string
	improvement string
	noteSummary string
	risk        int
}

// buildLineDecisionCopy menyusun analisa LINE dari campuran bukti nyata — bukan template seragam.
func buildLineDecisionCopy(line string, bucket *lineBucket, total int) lineDecisionCopy {
	idle := bucket.findings["idle"]
	noteLow := bucket.findings["note_low"]
	hanca := bucket.findings["loss_hanca"]
	broken := bucket.findings["loss_broken"]
	frequentBroken := bucket.findings["frequent_broken"]
	offline := bucket.findings["offline"]
	noLogin := bucket.findings["no_login"]

	parts := make([]string, 0, 6)
	actions := make([]string, 0, 6)
	noteHints := make([]string, 0, 4)
	decisionBits := make([]string, 0, 4)
	risk := 70

	if n := len(hanca); n > 0 {
		people := shortPeopleNames(hanca, 2)
		part := fmt.Sprintf("%d tunggu hanca/bahan", n)
		if len(people) > 0 {
			part += " (" + strings.Join(people, ", ") + ")"
		}
		parts = append(parts, part)
		actions = append(actions, shortStep(
			"Pastikan hanca/bahan untuk",
			people,
			fmt.Sprintf("%d mesin", n),
		))
		decisionBits = append(decisionBits, fmt.Sprintf("%d tunggu hanca — prioritaskan supply", n))
		risk += 12
		noteHints = append(noteHints, sampleOperatorHints(hanca, 2)...)
	}

	if n := len(broken); n > 0 {
		machines := shortMachineNames(broken, 2)
		part := fmt.Sprintf("%d mesin bermasalah", n)
		if len(machines) > 0 {
			part += " (" + strings.Join(machines, ", ") + ")"
		}
		parts = append(parts, part)
		actions = append(actions, shortStep("Kirim teknis ke", machines, fmt.Sprintf("%d mesin rusak", n)))
		decisionBits = append(decisionBits, fmt.Sprintf("%d mesin rusak — selesaikan teknis", n))
		risk += 14
		noteHints = append(noteHints, sampleOperatorHints(broken, 2)...)
	}

	if n := len(frequentBroken); n > 0 {
		machines := shortMachineNames(frequentBroken, 2)
		part := fmt.Sprintf("%d sering rusak", n)
		if len(machines) > 0 {
			part += " (" + strings.Join(machines, ", ") + ")"
		}
		parts = append(parts, part)
		who := strings.Join(orFallbackHints(machines, "mesin rawan"), " & ")
		actions = append(actions, fmt.Sprintf(
			"Preventive %s: cek teknis + siapkan spare sebelum shift berikutnya",
			who,
		))
		decisionBits = append(decisionBits, fmt.Sprintf(
			"%d mesin sering rusak — jadwalkan preventive (≥%dx/%dhari)",
			n,
			AiFrequentBrokenMin,
			AiBrokenWindowDays,
		))
		risk += 12
	}

	if n := len(offline); n > 0 {
		machines := shortMachineNames(offline, 2)
		part := fmt.Sprintf("%d offline", n)
		if len(machines) > 0 {
			part += " (" + strings.Join(machines, ", ") + ")"
		}
		parts = append(parts, part)
		actions = append(actions, shortStep("Cek power/jaringan", machines, "mesin offline"))
		decisionBits = append(decisionBits, fmt.Sprintf("%d offline — recovery fisik", n))
		risk += 8
	}

	if n := len(noteLow); n > 0 {
		people := shortPeopleNames(noteLow, 2)
		theme := summarizeNoteThemes(noteLow)
		part := fmt.Sprintf("%d sudah ada note", n)
		if len(people) > 0 {
			part += " (" + strings.Join(people, ", ") + ")"
		}
		parts = append(parts, part)
		if theme != "" && len(people) > 0 {
			actions = append(actions, fmt.Sprintf(
				"Cek %s soal %s",
				strings.Join(people, " & "),
				theme,
			))
			decisionBits = append(decisionBits, fmt.Sprintf(
				"%d note aktif (%s) — jangan minta isi ulang",
				n,
				theme,
			))
		} else if len(people) > 0 {
			actions = append(actions, fmt.Sprintf("Follow-up note %s", strings.Join(people, " & ")))
			decisionBits = append(decisionBits, fmt.Sprintf("%d sudah ada note — verifikasi statusnya", n))
		} else {
			actions = append(actions, "Follow-up note yang sudah tercatat")
			decisionBits = append(decisionBits, fmt.Sprintf("%d sudah ada note — verifikasi statusnya", n))
		}
		risk += 6
		noteHints = append(noteHints, sampleOperatorHints(noteLow, 2)...)
	}

	if n := len(idle); n > 0 {
		people := shortPeopleNames(idle, 2)
		if len(people) == 0 {
			people = shortMachineNames(idle, 2)
		}
		part := fmt.Sprintf("%d tanpa note", n)
		if len(people) > 0 {
			part += " (" + strings.Join(people, ", ") + ")"
		}
		parts = append(parts, part)
		actions = append(actions, shortStep("Minta isi note", people, "operator terkait"))
		decisionBits = append(decisionBits, fmt.Sprintf("%d masih tanpa note", n))
		risk += 10
	}

	if n := len(noLogin); n > 0 {
		machines := shortMachineNames(noLogin, 2)
		part := fmt.Sprintf("%d tanpa login (ON/jalan)", n)
		if len(machines) > 0 {
			part += " (" + strings.Join(machines, ", ") + ")"
		}
		parts = append(parts, part)
		actions = append(actions, shortStep("Login atau matikan", machines, "mesin ON/jalan tanpa login"))
		decisionBits = append(decisionBits, fmt.Sprintf("%d mesin ON/jalan tanpa login", n))
		risk += 10
	}

	if len(parts) == 0 {
		return lineDecisionCopy{}
	}
	if risk > 98 {
		risk = 98
	}
	risk += minInt(total, 5)

	evidence := strings.Join(parts, " · ")
	uniqNotes := uniqueStrings(noteHints)
	if len(uniqNotes) > 0 {
		evidence = evidence + " · Catatan: " + strings.Join(uniqNotes, "; ")
	}

	decision := fmt.Sprintf(
		"%s (%d mesin): %s",
		line,
		total,
		strings.Join(decisionBits, ". "),
	)
	if len(decisionBits) == 1 && len(idle) > 0 && len(noteLow) == 0 &&
		len(hanca) == 0 && len(broken) == 0 && len(offline) == 0 {
		decision = fmt.Sprintf(
			"%s: %d mesin util rendah tanpa note (%s) — lengkapi keterangan dulu sebelum menafsirkan utilisasi",
			line,
			len(idle),
			strings.Join(orFallbackHints(sampleMachineUtilHints(idle, 2), "util rendah"), "; "),
		)
	}

	// Langkah bernomor singkat — mudah dibaca di kartu.
	improvement := formatNumberedSteps(actions)
	prediction := fmt.Sprintf(
		"Jika langkah di atas ditunda, %s berisiko tetap tertekan 1 jam ke depan",
		line,
	)

	return lineDecisionCopy{
		signal:      fmt.Sprintf("%s: %d mesin perlu keputusan sekarang", line, total),
		evidence:    evidence,
		prediction:  prediction,
		decision:    decision,
		improvement: improvement,
		noteSummary: strings.Join(uniqNotes, "; "),
		risk:        risk,
	}
}

func orFallbackHints(hints []string, fallback string) []string {
	if len(hints) == 0 {
		return []string{fallback}
	}
	return hints
}

type frequentBrokenPreventive struct {
	decision    string
	improvement string
	prediction  string
}

// buildFrequentBrokenPreventive menyusun aksi pencegahan berbasis riwayat rusak (bukan spekulasi).
func buildFrequentBrokenPreventive(name, loc string, count int, activeNow bool) frequentBrokenPreventive {
	line := fallbackText(normalizeAiLineKey(loc), loc)
	where := fallbackText(line, "line")

	steps := make([]string, 0, 4)
	if activeNow {
		steps = append(steps, fmt.Sprintf("Selesaikan penanganan rusak aktif di %s dulu", name))
	}
	steps = append(steps,
		fmt.Sprintf("Jadwalkan cek teknis %s sebelum shift berikutnya (%s)", name, where),
		fmt.Sprintf("Siapkan spare/mekanik standby untuk %s (riwayat %dx/%dhari)", name, count, AiBrokenWindowDays),
	)
	if count >= 5 {
		steps = append(steps, fmt.Sprintf(
			"Eskalasi ke SPV/mekanik: %s rawan berulang — pertimbangkan rotasi proses kritis",
			name,
		))
	} else {
		steps = append(steps, fmt.Sprintf(
			"Pantau %s di awal shift; bila gejala ulang, tarik ke teknis segera",
			name,
		))
	}

	decision := fmt.Sprintf(
		"Preventive: %s sering rusak %dx/%dhari — cek teknis sebelum shift berikutnya",
		name,
		count,
		AiBrokenWindowDays,
	)
	if activeNow {
		decision = fmt.Sprintf(
			"%s rusak aktif + sering rusak %dx/%dhari — selesaikan sekarang, lalu preventive shift berikutnya",
			name,
			count,
			AiBrokenWindowDays,
		)
	}

	prediction := fmt.Sprintf(
		"Tanpa preventive, %s berisiko ulang rusak dan menekan utilitas %s 1 jam–shift ke depan",
		name,
		where,
	)
	if count >= 5 {
		prediction = fmt.Sprintf(
			"Pola rusak tinggi (%dx/%dhari): tanpa preventive, %s sangat mungkin ganggu line lagi shift berikutnya",
			count,
			AiBrokenWindowDays,
			name,
		)
	}

	return frequentBrokenPreventive{
		decision:    decision,
		improvement: formatNumberedSteps(steps),
		prediction:  prediction,
	}
}

func formatNumberedSteps(actions []string) string {
	if len(actions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(actions))
	for i, action := range actions {
		action = strings.TrimSpace(action)
		if action == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, action))
	}
	return strings.Join(parts, "\n")
}

func shortStep(verb string, names []string, fallback string) string {
	verb = strings.TrimSpace(verb)
	if len(names) == 0 {
		return strings.TrimSpace(verb + " " + fallback)
	}
	return strings.TrimSpace(verb + " " + strings.Join(names, " & "))
}

func ctxOperatorName(ctx aiMachineContext) string {
	if ctx.session != nil {
		if n := strings.TrimSpace(ctx.session.OperatorName); n != "" {
			return n
		}
	}
	if n := strings.TrimSpace(ctx.input.OperatorName); n != "" {
		return n
	}
	return ""
}

func ctxMachineName(ctx aiMachineContext) string {
	name := strings.TrimSpace(ctx.input.MachineName)
	if name == "" {
		name = strings.TrimSpace(ctx.input.UUID)
	}
	return clipAiText(name, 28)
}

func shortPeopleNames(list []aiMachineContext, limit int) []string {
	out := make([]string, 0, limit)
	seen := map[string]bool{}
	for _, ctx := range list {
		name := ctxOperatorName(ctx)
		if name == "" {
			name = ctxMachineName(ctx)
		}
		// Ambil nama pendek (kata terakhir sering nama panggilan, tapi aman pakai 2 kata terakhir).
		name = shortenPersonName(name)
		key := strings.ToUpper(name)
		if name == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, name)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func shortMachineNames(list []aiMachineContext, limit int) []string {
	out := make([]string, 0, limit)
	seen := map[string]bool{}
	for _, ctx := range list {
		name := ctxMachineName(ctx)
		key := strings.ToUpper(name)
		if name == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, name)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func shortenPersonName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	// Buang prefix NIK "92400934 - NAMA"
	if idx := strings.Index(name, " - "); idx >= 0 {
		name = strings.TrimSpace(name[idx+3:])
	}
	parts := strings.Fields(name)
	if len(parts) <= 2 {
		return strings.Join(parts, " ")
	}
	return strings.Join(parts[:2], " ")
}

func summarizeNoteThemes(list []aiMachineContext) string {
	themes := make([]string, 0, 3)
	seen := map[string]bool{}
	add := func(label string) {
		if label == "" || seen[label] {
			return
		}
		seen[label] = true
		themes = append(themes, label)
	}

	for _, ctx := range list {
		blob := strings.ToUpper(strings.TrimSpace(
			buildNoteSummary(ctx) + " " + ctx.input.OperatorNotes,
		))
		if blob == "" {
			continue
		}
		switch {
		case strings.Contains(blob, "HANCA") || strings.Contains(blob, "BAHAN") ||
			strings.Contains(blob, "AMBIL BAHAN") || strings.Contains(blob, "MATERIAL"):
			add("bahan/hanca")
		case strings.Contains(blob, "BENANG") || strings.Contains(blob, "GANTI BENANG"):
			add("ganti benang")
		case strings.Contains(blob, "MESIN RUSAK") || strings.Contains(blob, "BROKEN") ||
			strings.Contains(blob, "MACHINE_BROKEN"):
			add("teknis mesin")
		case strings.Contains(blob, "MEETING") || strings.Contains(blob, "RAPAT"):
			add("meeting/rapat")
		case strings.Contains(blob, "SOLAT") || strings.Contains(blob, "TOILET"):
			add("istirahat personal")
		default:
			if s := buildNoteSummary(ctx); s != "" {
				add(clipAiText(s, 36))
			}
		}
		if len(themes) >= 3 {
			break
		}
	}
	return strings.Join(themes, ", ")
}

func buildNoteSummary(ctx aiMachineContext) string {
	parts := make([]string, 0, 3)

	if ctx.loss != nil {
		label := strings.TrimSpace(ctx.loss.ReasonLabel)
		if label == "" {
			label = humanizeLossCode(ctx.loss.ReasonCode)
		}
		chunk := label
		if n := strings.TrimSpace(ctx.loss.Note); n != "" {
			chunk = label + " — " + clipAiText(n, 60)
		}
		parts = append(parts, chunk)
	}

	for _, note := range ctx.notes {
		label := strings.TrimSpace(note.ReasonName)
		if label == "" {
			label = humanizeLossCode(note.ReasonCode)
		}
		chunk := label
		if n := strings.TrimSpace(note.Note); n != "" {
			chunk = label + " — " + clipAiText(n, 60)
		}
		if chunk != "" {
			parts = append(parts, chunk)
		}
		if len(parts) >= 3 {
			break
		}
	}

	// Snapshot dashboard (termasuk loss event closed yang tampil di tabel).
	// Selalu digabung — jangan diabaikan hanya karena ada baris DB lain.
	if n := strings.TrimSpace(ctx.input.OperatorNotes); n != "" {
		parts = append(parts, clipAiText(n, 140))
	}

	return strings.Join(uniqueStrings(parts), " | ")
}

func sampleOperatorHints(list []aiMachineContext, limit int) []string {
	out := make([]string, 0, limit)
	for _, ctx := range list {
		name := ""
		if ctx.session != nil {
			name = strings.TrimSpace(ctx.session.OperatorName)
		}
		if name == "" {
			name = strings.TrimSpace(ctx.input.OperatorName)
		}
		if name == "" {
			name = strings.TrimSpace(ctx.input.MachineName)
		}
		hint := name
		if ctx.loss != nil {
			label := strings.TrimSpace(ctx.loss.ReasonLabel)
			if label == "" {
				label = humanizeLossCode(ctx.loss.ReasonCode)
			}
			hint = fmt.Sprintf("%s (%s %s)", name, label, formatAiDuration(ctx.loss.DurationSec))
			if n := strings.TrimSpace(ctx.loss.Note); n != "" {
				hint += " — " + clipAiText(n, 40)
			}
		} else if s := buildNoteSummary(ctx); s != "" {
			hint = fmt.Sprintf("%s — %s", name, clipAiText(s, 50))
		}
		out = append(out, hint)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// BuildAiInsightSummary narasi eksekutif untuk acuan keputusan.
func BuildAiInsightSummary(items []models.AiInsightItem, machineCount int) string {
	high, medium, low := 0, 0, 0
	for _, item := range items {
		switch item.Priority {
		case AiPriorityHigh:
			high++
		case AiPriorityMedium:
			medium++
		default:
			low++
		}
	}

	if len(items) == 0 {
		return fmt.Sprintf(
			"Ringkasan keputusan: dari %d mesin aktif yang dipantau, tidak ada sinyal risiko material untuk 1 jam ke depan. Pertahankan supervise rutin dan disiplin keterangan operator.",
			machineCount,
		)
	}

	// Ambil hingga 2 fokusasi line/mesin teratas.
	focusParts := make([]string, 0, 2)
	for _, item := range items {
		if len(focusParts) >= 2 {
			break
		}
		label := strings.TrimSpace(item.Location)
		if item.Scope != "LINE" {
			label = strings.TrimSpace(item.MachineName)
			if label == "" {
				label = strings.TrimSpace(item.Location)
			}
		}
		decision := strings.TrimSpace(item.Decision)
		if decision == "" {
			decision = strings.TrimSpace(item.Improvement)
		}
		focusParts = append(focusParts, fmt.Sprintf("%s → %s", label, trimAiSentence(decision)))
	}

	return fmt.Sprintf(
		"Ringkasan keputusan (%d mesin aktif): %d prioritas segera, %d perlu perhatian, %d info. Acuan utama: %s. Gunakan daftar di bawah sebagai urutan intervensi 1 jam ke depan.",
		machineCount,
		high,
		medium,
		low,
		strings.Join(focusParts, " | "),
	)
}

func humanizeLossCode(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "MACHINE_BROKEN":
		return "Mesin Rusak"
	case "WAIT_HANCA":
		return "Tunggu Hanca"
	case "TOILET":
		return "Toilet"
	case "PRAYER", "SOLAT":
		return "Solat"
	case "OTHER":
		return "Lainnya"
	case "":
		return "Keterangan"
	default:
		return strings.TrimSpace(code)
	}
}

func calcAiUtilPct(runtimeSec, procSec int64) float64 {
	if runtimeSec <= 0 {
		return 0
	}
	pct := float64(procSec) * 100.0 / float64(runtimeSec)
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return float64(int(pct*10+0.5)) / 10
}

// resolveAiUtilPct menyelaraskan angka AI dengan productivityPct di dashboard.
func resolveAiUtilPct(m models.AiInsightMachineInput) float64 {
	if m.ProductivityPct > 0 {
		pct := m.ProductivityPct
		if pct > 100 {
			pct = 100
		}
		return float64(int(pct*10+0.5)) / 10
	}
	return calcAiUtilPct(m.RuntimeSec, m.ProcSec)
}

func normalizeAiUUID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeAiMacState(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			return string(ch)
		}
	}
	upper := strings.ToUpper(value)
	switch upper {
	case "OFFLINE", "OFF":
		return "0"
	case "ONLINE", "ON", "IDLE":
		return "1"
	case "WORKING", "RUN", "RUNNING":
		return "2"
	default:
		return value
	}
}

func normalizeAiLineKey(location string) string {
	location = strings.TrimSpace(location)
	if location == "" || location == "-" {
		return ""
	}
	if idx := strings.Index(location, " - "); idx > 0 {
		location = strings.TrimSpace(location[:idx])
	}
	return strings.ToUpper(location)
}

func formatAiDuration(sec int64) string {
	if sec < 0 {
		sec = 0
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dj %dm", h, m)
	}
	if m == 0 && sec > 0 {
		return fmt.Sprintf("%d dtk", sec)
	}
	return fmt.Sprintf("%d menit", m)
}

func trimAiSentence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	return value
}

func clipAiText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return strings.TrimSpace(value[:max]) + "…"
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func joinAiParts(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " · ")
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
