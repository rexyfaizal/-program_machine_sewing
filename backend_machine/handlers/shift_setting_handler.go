package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (h *Handler) ShiftSettings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		h.getShiftSettings(ctx, w, r)
	case http.MethodPut, http.MethodPost:
		h.putShiftSettings(ctx, w, r)
	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getShiftSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	area := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("area")))
	if area == "" {
		area = strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("factory")))
	}
	if area == "" {
		http.Error(w, "area/factory wajib diisi", http.StatusBadRequest)
		return
	}

	allShifts, err := h.Repo.ListShiftSettingsForArea(ctx, area)
	if err != nil {
		http.Error(w, "Gagal ambil shift_setting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pisah jadwal default area (line_name='') dari override per line.
	defaultShifts := make([]models.ShiftSetting, 0)
	overrideByLine := map[string][]models.ShiftSetting{}
	for _, s := range allShifts {
		s.StartTime = trimClockHHMM(s.StartTime)
		s.EndTime = trimClockHHMM(s.EndTime)
		s.BreakStart = trimClockHHMM(s.BreakStart)
		s.BreakEnd = trimClockHHMM(s.BreakEnd)
		line := strings.ToUpper(strings.TrimSpace(s.LineName))
		if line == "" {
			defaultShifts = append(defaultShifts, s)
		} else {
			overrideByLine[line] = append(overrideByLine[line], s)
		}
	}

	lineConfigs, err := h.Repo.GetLineShiftConfigs(ctx, area)
	if err != nil {
		http.Error(w, "Gagal ambil line shift config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lines := make([]models.ShiftSettingLineInput, 0, len(lineConfigs))
	for _, cfg := range lineConfigs {
		lineKey := strings.ToUpper(strings.TrimSpace(cfg.LineName))
		override, hasOverride := overrideByLine[lineKey]
		item := models.ShiftSettingLineInput{
			LineName: cfg.LineName,
			Enabled:  cfg.Enabled,
			Custom:   hasOverride,
		}
		if hasOverride {
			item.Shifts = override
		}
		lines = append(lines, item)
	}

	writeJSON(w, map[string]any{
		"area":   area,
		"shifts": defaultShifts,
		"lines":  lines,
		"defaults": map[string]any{
			"schedule": utils.DefaultGM3ScheduleItems(),
		},
	})
}

func (h *Handler) putShiftSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var input models.ShiftSettingPutRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	area := strings.ToUpper(strings.TrimSpace(input.Area))
	if area == "" {
		http.Error(w, "area wajib diisi", http.StatusBadRequest)
		return
	}

	defaultShifts, err := normalizeShiftInputs(input.Shifts, area, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defaultSchedule := shiftsToSchedule(defaultShifts)

	overrides := map[string][]models.ShiftSetting{}
	linePayload := make([]models.LineShiftConfig, 0, len(input.Lines))

	for _, line := range input.Lines {
		lineName := strings.TrimSpace(line.LineName)
		if lineName == "" {
			continue
		}
		lineKey := strings.ToUpper(lineName)

		cfg := models.LineShiftConfig{
			Factory:  area,
			LineName: lineName,
			Enabled:  line.Enabled,
		}

		if line.Enabled && line.Custom {
			lineShifts, err := normalizeShiftInputs(line.Shifts, area, lineKey)
			if err != nil {
				http.Error(w, "Line "+lineName+": "+err.Error(), http.StatusBadRequest)
				return
			}
			overrides[lineKey] = lineShifts
			cfg.Schedule = shiftsToSchedule(lineShifts)
		} else if line.Enabled {
			// Pakai jadwal default area.
			if len(defaultShifts) == 0 {
				http.Error(w, "Line "+lineName+" mode Shift tapi jadwal default area kosong. Tambah jadwal default atau custom, atau set line ke Normal.", http.StatusBadRequest)
				return
			}
			cfg.Schedule = defaultSchedule
		}

		linePayload = append(linePayload, cfg)
	}

	if err := h.Repo.ReplaceShiftSettings(ctx, area, defaultShifts, overrides); err != nil {
		http.Error(w, "Gagal simpan shift_setting: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.Repo.UpsertLineShiftConfigs(ctx, area, linePayload); err != nil {
		http.Error(w, "Gagal sync line shift config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// normalizeShiftInputs memvalidasi + membersihkan daftar shift.
func normalizeShiftInputs(shifts []models.ShiftSetting, area, lineName string) ([]models.ShiftSetting, error) {
	out := make([]models.ShiftSetting, 0, len(shifts))
	for i, shift := range shifts {
		name := strings.ToUpper(strings.TrimSpace(shift.ShiftName))
		if name == "" {
			name = fmt.Sprintf("SHIFT_%d", i+1)
		}

		start := strings.TrimSpace(shift.StartTime)
		end := strings.TrimSpace(shift.EndTime)
		if start == "" || end == "" {
			return nil, fmt.Errorf("jam mulai/selesai wajib untuk %s", name)
		}
		if _, err := utils.ParseClockToMinutes(start); err != nil {
			return nil, fmt.Errorf("jam mulai tidak valid: %s", name)
		}
		if _, err := utils.ParseClockToMinutes(end); err != nil {
			return nil, fmt.Errorf("jam selesai tidak valid: %s", name)
		}

		shiftNo := shift.ShiftNo
		if shiftNo <= 0 {
			shiftNo = i + 1
		}

		out = append(out, models.ShiftSetting{
			Area:       area,
			LineName:   lineName,
			ShiftNo:    shiftNo,
			ShiftName:  name,
			StartTime:  start,
			EndTime:    end,
			BreakStart: strings.TrimSpace(shift.BreakStart),
			BreakEnd:   strings.TrimSpace(shift.BreakEnd),
			IsActive:   true,
		})
	}

	if len(out) > 0 {
		schedule := shiftsToSchedule(out)
		if len(utils.ScheduleToSegments(schedule)) == 0 {
			return nil, fmt.Errorf("jadwal shift tidak menghasilkan jam kerja valid")
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].ShiftNo < out[j].ShiftNo })
	return out, nil
}

func shiftsToSchedule(shifts []models.ShiftSetting) []models.ShiftScheduleItem {
	schedule := make([]models.ShiftScheduleItem, 0, len(shifts))
	for _, s := range shifts {
		schedule = append(schedule, models.ShiftScheduleItem{
			Code:       s.ShiftName,
			Start:      trimClockHHMM(s.StartTime),
			End:        trimClockHHMM(s.EndTime),
			BreakStart: trimClockHHMM(s.BreakStart),
			BreakEnd:   trimClockHHMM(s.BreakEnd),
		})
	}
	return schedule
}

func trimClockHHMM(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}
