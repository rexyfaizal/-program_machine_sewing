package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend_machine/models"
)

func (h *Handler) MachineSettings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		data, err := h.Repo.GetMachineSettings(ctx)
		if err != nil {
			http.Error(w, "Gagal ambil machine setting: "+err.Error(), http.StatusInternalServerError)
			return
		}

		list := make([]models.MachineSetting, 0)
		for _, item := range data {
			list = append(list, item)
		}

		writeJSON(w, list)

	case http.MethodPost, http.MethodPut:
		var input models.MachineSetting

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
			return
		}

		input.UUID = strings.TrimSpace(input.UUID)
		input.CustomName = strings.TrimSpace(input.CustomName)
		input.Location = strings.TrimSpace(input.Location)
		input.Pic = strings.TrimSpace(input.Pic)
		input.Spv = strings.TrimSpace(input.Spv)

		if input.UUID == "" {
			http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
			return
		}

		if err := h.Repo.UpsertMachineSetting(ctx, input); err != nil {
			http.Error(w, "Gagal simpan machine setting: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]string{
			"status": "ok",
		})

	case http.MethodDelete:
		uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
		if uuid == "" {
			http.Error(w, "uuid wajib diisi", http.StatusBadRequest)
			return
		}

		if err := h.Repo.DeleteMachineSetting(ctx, uuid); err != nil {
			http.Error(w, "Gagal hapus machine setting: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]string{
			"status": "ok",
		})

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}
