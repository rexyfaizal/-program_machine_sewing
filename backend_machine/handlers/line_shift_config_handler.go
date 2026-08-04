package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

func (h *Handler) LineShiftConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		factory := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("factory")))
		list, err := h.Repo.GetLineShiftConfigs(ctx, factory)
		if err != nil {
			http.Error(w, "Gagal ambil line shift config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]any{
			"factory": factory,
			"lines":   list,
			"defaults": map[string]any{
				"schedule": utils.DefaultGM3ScheduleItems(),
			},
		})

	case http.MethodPut, http.MethodPost:
		var input models.LineShiftConfigPutRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
			return
		}

		factory := strings.ToUpper(strings.TrimSpace(input.Factory))
		if factory == "" {
			http.Error(w, "factory wajib diisi", http.StatusBadRequest)
			return
		}

		for i := range input.Lines {
			input.Lines[i].Factory = factory
			input.Lines[i].LineName = strings.TrimSpace(input.Lines[i].LineName)
			if input.Lines[i].LineName == "" {
				http.Error(w, "lineName wajib diisi", http.StatusBadRequest)
				return
			}

			if input.Lines[i].Enabled {
				if len(input.Lines[i].Schedule) == 0 {
					http.Error(w, "schedule wajib diisi jika shift aktif: "+input.Lines[i].LineName, http.StatusBadRequest)
					return
				}

				for _, item := range input.Lines[i].Schedule {
					code := utils.NormalizeShiftCode(item.Code)
					if code == utils.ShiftALL || code == utils.ShiftCurrent {
						http.Error(w, "kode shift tidak valid", http.StatusBadRequest)
						return
					}
					if _, err := utils.ParseClockToMinutes(item.Start); err != nil {
						http.Error(w, "jam mulai tidak valid pada "+input.Lines[i].LineName, http.StatusBadRequest)
						return
					}
					if _, err := utils.ParseClockToMinutes(item.End); err != nil {
						http.Error(w, "jam selesai tidak valid pada "+input.Lines[i].LineName, http.StatusBadRequest)
						return
					}
				}

				if len(utils.ScheduleToSegments(input.Lines[i].Schedule)) == 0 {
					http.Error(w, "schedule tidak menghasilkan jam kerja valid: "+input.Lines[i].LineName, http.StatusBadRequest)
					return
				}
			}
		}

		if err := h.Repo.UpsertLineShiftConfigs(ctx, factory, input.Lines); err != nil {
			http.Error(w, "Gagal simpan line shift config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]string{"status": "ok"})

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}
