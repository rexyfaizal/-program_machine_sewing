package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend_machine/models"
)

const (
	ShiftALL     = "ALL"
	ShiftCurrent = "CURRENT"
	Shift1       = "SHIFT_1"
	Shift2       = "SHIFT_2"
	Shift3       = "SHIFT_3"

	GM3BreakMinutes = 30
)

// ShiftSegment adalah potongan jam kerja aktif (sudah exclude istirahat).
type ShiftSegment struct {
	ShiftNo   int
	ShiftName string
	StartMin  int // menit dari work_date 00:00
	EndMin    int
}

// GM3ShiftSegments mendefinisikan 3 shift GM3 default:
// SHIFT_1 06:00–13:30 istirahat 09:30–10:00
// SHIFT_2 13:30–21:00 istirahat 17:00–17:30
// SHIFT_3 21:00–04:30(+1) istirahat 01:00–01:30
func GM3ShiftSegments() []ShiftSegment {
	breakMin := GM3BreakMinutes

	return []ShiftSegment{
		{1, Shift1, 360, 570},
		{1, Shift1, 570 + breakMin, 810},
		{2, Shift2, 810, 1020},
		{2, Shift2, 1020 + breakMin, 1260},
		{3, Shift3, 1260, 1500},
		{3, Shift3, 1500 + breakMin, 1710},
	}
}

func DefaultGM3ScheduleItems() []models.ShiftScheduleItem {
	return []models.ShiftScheduleItem{
		{Code: Shift1, Start: "06:00", End: "13:30", BreakStart: "09:30", BreakEnd: "10:00"},
		{Code: Shift2, Start: "13:30", End: "21:00", BreakStart: "17:00", BreakEnd: "17:30"},
		{Code: Shift3, Start: "21:00", End: "04:30", BreakStart: "01:00", BreakEnd: "01:30"},
	}
}

func NormalizeShiftCode(shift string) string {
	code := strings.ToUpper(strings.TrimSpace(shift))

	switch code {
	case "", ShiftALL:
		return ShiftALL
	case ShiftCurrent, "NOW":
		return ShiftCurrent
	case Shift1, "1", "S1":
		return Shift1
	case Shift2, "2", "S2":
		return Shift2
	case Shift3, "3", "S3":
		return Shift3
	default:
		return ShiftALL
	}
}

func ShiftDisplayName(code string) string {
	switch NormalizeShiftCode(code) {
	case Shift1:
		return "Shift 1 (06:00-13:30)"
	case Shift2:
		return "Shift 2 (13:30-21:00)"
	case Shift3:
		return "Shift 3 (21:00-04:30)"
	case ShiftCurrent:
		return "Current Shift"
	default:
		return "All Shifts (06:00-04:30)"
	}
}

func ShiftDisplayNameFromSchedule(code string, schedule []models.ShiftScheduleItem) string {
	normalized := NormalizeShiftCode(code)
	if normalized == ShiftCurrent {
		return "Current Shift"
	}
	if normalized == ShiftALL {
		return "All Shifts"
	}

	for _, item := range schedule {
		if NormalizeShiftCode(item.Code) == normalized {
			start := strings.TrimSpace(item.Start)
			end := strings.TrimSpace(item.End)
			if start != "" && end != "" {
				return fmt.Sprintf("Shift %d (%s-%s)", ShiftNumber(normalized), start, end)
			}
		}
	}

	return ShiftDisplayName(normalized)
}

func ShiftNumber(code string) int {
	switch NormalizeShiftCode(code) {
	case Shift1:
		return 1
	case Shift2:
		return 2
	case Shift3:
		return 3
	default:
		return 0 // ALL / CURRENT unresolved
	}
}

// ParseClockToMinutes mengubah "HH:MM" / "H:MM" menjadi menit dari midnight.
func ParseClockToMinutes(clock string) (int, error) {
	text := strings.TrimSpace(clock)
	if text == "" {
		return 0, fmt.Errorf("jam kosong")
	}

	parts := strings.Split(text, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("format jam tidak valid: %s", clock)
	}

	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("jam tidak valid: %s", clock)
	}

	minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("menit tidak valid: %s", clock)
	}

	return hour*60 + minute, nil
}

func ParseShiftScheduleJSON(raw string) []models.ShiftScheduleItem {
	text := strings.TrimSpace(raw)
	if text == "" {
		return []models.ShiftScheduleItem{}
	}

	var items []models.ShiftScheduleItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return []models.ShiftScheduleItem{}
	}

	result := make([]models.ShiftScheduleItem, 0, len(items))
	for _, item := range items {
		code := NormalizeShiftCode(item.Code)
		if code == ShiftALL || code == ShiftCurrent {
			continue
		}
		result = append(result, models.ShiftScheduleItem{
			Code:       code,
			Start:      strings.TrimSpace(item.Start),
			End:        strings.TrimSpace(item.End),
			BreakStart: strings.TrimSpace(item.BreakStart),
			BreakEnd:   strings.TrimSpace(item.BreakEnd),
		})
	}
	return result
}

// ScheduleToSegments mengubah jadwal admin menjadi segment kerja (exclude break).
func ScheduleToSegments(schedule []models.ShiftScheduleItem) []ShiftSegment {
	segments := make([]ShiftSegment, 0)

	for _, item := range schedule {
		code := NormalizeShiftCode(item.Code)
		shiftNo := ShiftNumber(code)
		if shiftNo == 0 {
			continue
		}

		startMin, err := ParseClockToMinutes(item.Start)
		if err != nil {
			continue
		}
		endMin, err := ParseClockToMinutes(item.End)
		if err != nil {
			continue
		}

		// Lintas midnight (mis. 21:00–04:30) → end di hari berikutnya.
		if endMin <= startMin {
			endMin += 24 * 60
		}

		breakStart, breakEnd := -1, -1
		if bs, err1 := ParseClockToMinutes(item.BreakStart); err1 == nil {
			if be, err2 := ParseClockToMinutes(item.BreakEnd); err2 == nil {
				breakStart = bs
				breakEnd = be
				if breakStart < startMin {
					breakStart += 24 * 60
				}
				if breakEnd <= breakStart {
					breakEnd += 24 * 60
				}
				if breakEnd > endMin {
					breakEnd = endMin
				}
				if breakStart < startMin || breakStart >= endMin {
					breakStart, breakEnd = -1, -1
				}
			}
		}

		if breakStart < 0 || breakEnd < 0 || breakEnd <= breakStart {
			segments = append(segments, ShiftSegment{
				ShiftNo:   shiftNo,
				ShiftName: code,
				StartMin:  startMin,
				EndMin:    endMin,
			})
			continue
		}

		if breakStart > startMin {
			segments = append(segments, ShiftSegment{
				ShiftNo:   shiftNo,
				ShiftName: code,
				StartMin:  startMin,
				EndMin:    breakStart,
			})
		}
		if breakEnd < endMin {
			segments = append(segments, ShiftSegment{
				ShiftNo:   shiftNo,
				ShiftName: code,
				StartMin:  breakEnd,
				EndMin:    endMin,
			})
		}
	}

	return segments
}

func FilterSegmentsByShift(segments []ShiftSegment, shiftCode string) []ShiftSegment {
	code := NormalizeShiftCode(shiftCode)
	if code == ShiftALL || code == ShiftCurrent || ShiftNumber(code) == 0 {
		return segments
	}

	filtered := make([]ShiftSegment, 0)
	for _, seg := range segments {
		if seg.ShiftName == code {
			filtered = append(filtered, seg)
		}
	}
	return filtered
}

func SegmentBounds(segments []ShiftSegment) (workStartMin, workEndMin int) {
	if len(segments) == 0 {
		return 0, 24 * 60
	}

	workStartMin = segments[0].StartMin
	workEndMin = segments[0].EndMin
	for _, seg := range segments {
		if seg.StartMin < workStartMin {
			workStartMin = seg.StartMin
		}
		if seg.EndMin > workEndMin {
			workEndMin = seg.EndMin
		}
	}
	return workStartMin, workEndMin
}

// ResolveShiftSegmentsForLocation menentukan segment shift untuk mesin.
// - Config enabled → pakai schedule line
// - Config disabled → full day
// - Tanpa config + GM3 → default GM3 (legacy)
// - Tanpa config selain GM3 → full day
func ResolveShiftSegmentsForLocation(
	location string,
	configMap map[string]models.LineShiftConfig,
) (useShift bool, segments []ShiftSegment, schedule []models.ShiftScheduleItem) {
	factory, lineName := ParseLocationParts(location)

	if factory != "" && lineName != "" && configMap != nil {
		if cfg, ok := configMap[LineShiftConfigKey(factory, lineName)]; ok {
			if !cfg.Enabled {
				return false, nil, nil
			}
			schedule = cfg.Schedule
			if len(schedule) == 0 {
				schedule = DefaultGM3ScheduleItems()
			}
			segments = ScheduleToSegments(schedule)
			if len(segments) == 0 {
				return false, nil, nil
			}
			return true, segments, schedule
		}
	}

	if IsGM3Location(location) {
		return true, GM3ShiftSegments(), DefaultGM3ScheduleItems()
	}

	return false, nil, nil
}

func FactoryHasEnabledShift(factory string, configMap map[string]models.LineShiftConfig) bool {
	factory = strings.ToUpper(strings.TrimSpace(factory))
	if factory == "" {
		return false
	}

	if configMap != nil {
		found := false
		for _, cfg := range configMap {
			if strings.ToUpper(strings.TrimSpace(cfg.Factory)) != factory {
				continue
			}
			found = true
			if cfg.Enabled && len(cfg.Schedule) > 0 {
				return true
			}
		}
		if found {
			return false
		}
	}

	// Belum ada config tersimpan: GM3 tetap multi-shift (legacy).
	return factory == "GM3"
}

// ResolveGM3WorkDate menentukan work_date untuk GM3.
// Jam 00:00–04:29 masih termasuk work_date hari sebelumnya (SHIFT_3).
func ResolveGM3WorkDate(now time.Time) time.Time {
	local := now
	minutes := local.Hour()*60 + local.Minute()

	if minutes < 270 { // sebelum 04:30
		local = local.AddDate(0, 0, -1)
	}

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

// ResolveGM3WorkWindow mengembalikan rentang waktu hari kerja GM3 aktif.
// Dari 06:00 work_date sampai 04:30 hari kalender berikutnya.
func ResolveGM3WorkWindow(now time.Time) (time.Time, time.Time) {
	workDate := ResolveGM3WorkDate(now)
	loc := workDate.Location()

	start := time.Date(workDate.Year(), workDate.Month(), workDate.Day(), 6, 0, 0, 0, loc)
	end := start.Add(22*time.Hour + 30*time.Minute)

	return start, end
}

// ResolveGM3CurrentShift mengembalikan shift aktif berdasarkan waktu sekarang.
func ResolveGM3CurrentShift(now time.Time) (workDate time.Time, shiftCode string) {
	workDate = ResolveGM3WorkDate(now)
	base := workDate

	minutesFromWorkDate := int(now.Sub(base).Minutes())

	switch {
	case minutesFromWorkDate >= 360 && minutesFromWorkDate < 810:
		return workDate, Shift1
	case minutesFromWorkDate >= 810 && minutesFromWorkDate < 1260:
		return workDate, Shift2
	case minutesFromWorkDate >= 1260 && minutesFromWorkDate < 1710:
		return workDate, Shift3
	default:
		return workDate, ShiftALL
	}
}

// ResolveRequestedShift menyelesaikan CURRENT dan mengembalikan work_date efektif.
func ResolveRequestedShift(dateText string, shift string, now time.Time) (workDate string, shiftCode string) {
	shiftCode = NormalizeShiftCode(shift)
	workDate = strings.TrimSpace(dateText)

	if workDate == "" {
		workDate = now.Format("2006-01-02")
	}

	if shiftCode != ShiftCurrent {
		return workDate, shiftCode
	}

	resolvedDate, resolvedShift := ResolveGM3CurrentShift(now)
	todayWork := ResolveGM3WorkDate(now).Format("2006-01-02")

	if workDate != todayWork && workDate != resolvedDate.Format("2006-01-02") {
		return workDate, ShiftALL
	}

	return resolvedDate.Format("2006-01-02"), resolvedShift
}

func EffectiveWorkMinutes(shiftCode string) int {
	return EffectiveWorkMinutesFromSegments(GM3ShiftSegments(), shiftCode)
}

func EffectiveWorkMinutesFromSegments(segments []ShiftSegment, shiftCode string) int {
	code := NormalizeShiftCode(shiftCode)
	total := 0

	for _, seg := range segments {
		if code != ShiftALL && seg.ShiftName != code {
			continue
		}
		total += seg.EndMin - seg.StartMin
	}

	return total
}

// BuildShiftSegmentValuesSQL membangun baris VALUES aman dari segment tervalidasi.
func BuildShiftSegmentValuesSQL(segments []ShiftSegment) string {
	if len(segments) == 0 {
		return ""
	}

	parts := make([]string, 0, len(segments))
	for _, seg := range segments {
		name := NormalizeShiftCode(seg.ShiftName)
		if name == ShiftALL || name == ShiftCurrent {
			continue
		}
		parts = append(parts, fmt.Sprintf(
			"(%d, N'%s', DATEADD(MINUTE, %d, @base), DATEADD(MINUTE, %d, @base))",
			seg.ShiftNo,
			name,
			seg.StartMin,
			seg.EndMin,
		))
	}

	return strings.Join(parts, ",\n        ")
}
