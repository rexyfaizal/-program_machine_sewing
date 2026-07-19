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

	ctx, cancel := contextWithTimeout(r, 60*time.Second)
	defer cancel()

	machines, err := h.Repo.GetMachines(ctx)
	if err != nil {
		log.Println("Gagal ambil data machineinfo:", err)
		http.Error(w, "Gagal ambil data machineinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Tanggal request:", date)
	log.Println("Jumlah mesin dari machineinfo:", len(machines))

	rows := make([]models.ProductivityRow, 0)

	var sumPct float64
	var totalOutput int64
	var totalAlarm int64
	var totalProc int64
	var totalRun int64
	var good, normal, bad int

	for _, m := range machines {
		row, err := h.Repo.GetMachineProductivity(ctx, m, date)
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

	// Merge data manual dari machine_setting_manual:
	// custom_name, location, pic, spv.
	settings, err := h.Repo.GetMachineSettings(ctx)
	if err != nil {
		log.Println("Gagal ambil machine setting manual:", err)
	} else {
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
	}

	// Simpan / update snapshot harian ke logs_machine.
	// Logic di repository:
	// - Jika log_date + uuid belum ada  => INSERT
	// - Jika log_date + uuid sudah ada  => UPDATE
	// - Jika ganti hari                 => INSERT row baru
	// - Jika uuid mesin baru            => INSERT row baru
	// if err := h.Repo.SaveLogsMachine(ctx, date, rows); err != nil {
	// 	log.Println("Gagal simpan logs_machine:", err)
	// }

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

	resp := models.APIResponse{
		Summary: models.Summary{
			Date:        date,
			Total:       len(rows),
			Good:        good,
			Normal:      normal,
			Bad:         bad,
			AvgPct:      avg,
			WorkHours:   8,
			TotalOutput: totalOutput,
			TotalAlarm:  totalAlarm,
			TotalProc:   totalProc,
			TotalRun:    totalRun,
		},
		Rows: rows,
	}

	writeJSON(w, resp)
}
