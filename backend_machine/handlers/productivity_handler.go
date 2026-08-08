package handlers

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (h *Handler) Productivity(w http.ResponseWriter, r *http.Request) {
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	shiftParam := strings.TrimSpace(r.URL.Query().Get("shift"))
	workDate, shiftCode := utils.ResolveRequestedShift(date, shiftParam, time.Now())

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	machines, err := h.Repo.GetMachines(ctx)
	if err != nil {
		log.Println("Gagal ambil data machineinfo:", err)
		http.Error(w, "Gagal ambil data machineinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.Repo.GetMachineSettings(ctx)
	if err != nil {
		log.Println("Gagal ambil machine setting manual:", err)
		settings = map[string]models.MachineSetting{}
	}

	shiftConfigMap, err := h.Repo.GetLineShiftConfigMap(ctx)
	if err != nil {
		log.Println("Gagal ambil line shift config:", err)
		shiftConfigMap = map[string]models.LineShiftConfig{}
	}

	log.Println("Tanggal request:", date, "workDate:", workDate, "shift:", shiftCode)
	log.Println("Jumlah mesin dari machineinfo:", len(machines))

	rows := make([]models.ProductivityRow, 0)

	var sumPct float64
	var totalOutput int64
	var totalAlarm int64
	var totalProc int64
	var totalRun int64
	var good, normal, bad int

	for _, m := range machines {
		key := strings.ToLower(strings.TrimSpace(m.UUID))
		location := ""
		if setting, ok := settings[key]; ok {
			location = setting.Location
		}

		var row models.ProductivityRow
		var err error

		useShift, segments, schedule := utils.ResolveShiftSegmentsForLocation(location, shiftConfigMap)
		area, lineName := utils.ParseLocationParts(location)
		rawShift := strings.ToUpper(strings.TrimSpace(shiftParam))
		isFullDayNormal := rawShift == utils.ShiftNormal ||
			rawShift == "FULLDAY" ||
			rawShift == "FULL_DAY" ||
			rawShift == "HARI_PENUH" ||
			!useShift

		hasShiftSetting := false
		if area != "" {
			hasShiftSetting, err = h.Repo.HasActiveShiftSettings(ctx, area, workDate)
			if err != nil {
				log.Printf("shift_setting skip area=%s: %v", area, err)
				hasShiftSetting = false
				err = nil
			}
		}

		if isFullDayNormal {
			// Mode NORMAL (hari penuh) — rumus FINAL.
			row, err = h.Repo.GetMachineProductivityFinal(ctx, m, date, "NORMAL", area, lineName, "")
		} else if useShift && hasShiftSetting {
			// Mode SHIFT — jadwal dari dbo.shift_setting (line override → default area).
			code := shiftCode
			if shiftParam == "" {
				_, code = utils.ResolveRequestedShift(date, utils.ShiftCurrent, time.Now())
			}
			if code == "" || code == utils.ShiftCurrent {
				code = utils.ShiftALL
			}
			selected := code
			if code == utils.ShiftALL {
				selected = ""
			}
			row, err = h.Repo.GetMachineProductivityFinal(ctx, m, workDate, "SHIFT", area, lineName, selected)
		} else if useShift {
			// Fallback legacy (belum ada shift_setting).
			code := shiftCode
			if shiftParam == "" {
				_, code = utils.ResolveRequestedShift(date, utils.ShiftCurrent, time.Now())
			}
			if code == "" || code == utils.ShiftCurrent {
				code = utils.ShiftALL
			}
			row, err = h.Repo.GetMachineProductivityByShift(ctx, m, workDate, code, segments, schedule)
		} else {
			row, err = h.Repo.GetMachineProductivityFinal(ctx, m, date, "NORMAL", area, lineName, "")
		}

		if err != nil {
			log.Printf("skip machine table=%s uuid=%s: %v", m.TableName, m.UUID, err)
			continue
		}

		rows = append(rows, row)

		sumPct += row.ProductivityPct
		totalOutput += row.Output
		totalAlarm += row.AlarmCount
		totalProc += row.ProcSec
		totalRun += row.RuntimeSec

		switch row.Status {
		case "GOOD":
			good++
		case "NORMAL":
			normal++
		default:
			bad++
		}
	}

	for i := range rows {
		rows[i].OriginalNickName = rows[i].NickName

		key := strings.ToLower(strings.TrimSpace(rows[i].UUID))

		if setting, ok := settings[key]; ok {
			if setting.CustomName != "" {
				rows[i].NickName = setting.CustomName
				rows[i].MachineName = setting.CustomName
			}

			rows[i].Location = setting.Location
			rows[i].Pic = setting.Pic
			rows[i].Spv = setting.Spv
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ProductivityPct == rows[j].ProductivityPct {
			return rows[i].NickName < rows[j].NickName
		}
		return rows[i].ProductivityPct > rows[j].ProductivityPct
	})

	avg := 0.0
	if len(rows) > 0 {
		avg = utils.Round2(sumPct / float64(len(rows)))
	}

	workHours := 8.0
	if shiftCode != "" && shiftCode != utils.ShiftALL {
		mins := utils.EffectiveWorkMinutes(shiftCode)
		if mins > 0 {
			workHours = utils.Round2(float64(mins) / 60)
		}
	}

	resp := models.APIResponse{
		Summary: models.Summary{
			Date:        date,
			Total:       len(rows),
			Good:        good,
			Normal:      normal,
			Bad:         bad,
			AvgPct:      avg,
			WorkHours:   workHours,
			TotalOutput: totalOutput,
			TotalAlarm:  totalAlarm,
			TotalProc:   totalProc,
			TotalRun:    totalRun,
			ShiftCode:   shiftCode,
			ShiftName:   utils.ShiftDisplayName(shiftCode),
		},
		Rows: rows,
	}

	writeJSON(w, resp)
}
