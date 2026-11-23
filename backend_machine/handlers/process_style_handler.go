package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/repository"
	"backend_machine/utils"
)

func normalizeProcessStyleInput(input models.ProcessStyleRequest) models.ProcessStyleRequest {
	input.ProcessName = strings.TrimSpace(input.ProcessName)
	input.StyleName = strings.TrimSpace(input.StyleName)
	input.Proses = strings.TrimSpace(input.Proses)
	input.Style = strings.TrimSpace(input.Style)

	if input.ProcessName == "" {
		input.ProcessName = input.Proses
	}

	if input.StyleName == "" {
		input.StyleName = input.Style
	}

	return input
}

func getProcessStyleIDFromPath(path string) (int64, error) {
	prefix := "/api/process-style/"
	idText := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	idText = strings.Trim(idText, "/")

	if idText == "" {
		return 0, errors.New("id wajib diisi")
	}

	return strconv.ParseInt(idText, 10, 64)
}

// ProcessStyleStyles godoc
// @Summary Cari daftar style
// @Description Mengambil daftar style dari tabel gcc_smv_process
// @Tags Process Style
// @Produce json
// @Param q query string false "Keyword style"
// @Success 200 {array} models.ProcessStyleItem
// @Failure 500 {string} string
// @Router /process-style/styles [get]
func (h *Handler) ProcessStyleStyles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	data, err := h.Repo.GetProcessStyleStyles(ctx, q)
	if err != nil {
		http.Error(w, "Gagal ambil data style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

// ProcessStyleProcesses godoc
// @Summary Cari proses berdasarkan style
// @Description Mengambil daftar proses berdasarkan style dari tabel gcc_smv_process
// @Tags Process Style
// @Produce json
// @Param style query string true "Style"
// @Param q query string false "Keyword proses"
// @Success 200 {array} models.ProcessStyleProcess
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /process-style/processes [get]
func (h *Handler) ProcessStyleProcesses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	style := strings.TrimSpace(r.URL.Query().Get("style"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	if style == "" {
		http.Error(w, "style wajib diisi", http.StatusBadRequest)
		return
	}

	data, err := h.Repo.GetProcessStyleProcesses(ctx, style, q)
	if err != nil {
		http.Error(w, "Gagal ambil data proses style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

// ProcessStyleList godoc
// @Summary List proses style
// @Description Menampilkan semua proses style dengan filter q
// @Tags Process Style
// @Produce json
// @Param q query string false "Keyword style atau proses"
// @Success 200 {array} models.ProcessStyleRecord
// @Failure 500 {string} string
// @Router /process-style/list [get]
func (h *Handler) ProcessStyleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	q := strings.TrimSpace(r.URL.Query().Get("q"))

	data, err := h.Repo.ListProcessStyle(ctx, q)
	if err != nil {
		http.Error(w, "Gagal ambil list process style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

// ProcessStyleCreate godoc
// @Summary Tambah proses style
// @Description Menambahkan data proses dan style baru
// @Tags Process Style
// @Accept json
// @Produce json
// @Param body body models.ProcessStyleRequest true "Body process style"
// @Success 200 {object} models.ProcessStyleRecord
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /process-style [post]
func (h *Handler) ProcessStyleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	var input models.ProcessStyleRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	input = normalizeProcessStyleInput(input)

	if input.ProcessName == "" {
		http.Error(w, "processName/proses wajib diisi", http.StatusBadRequest)
		return
	}

	if input.StyleName == "" {
		http.Error(w, "styleName/style wajib diisi", http.StatusBadRequest)
		return
	}

	data, err := h.Repo.CreateProcessStyle(ctx, input)
	if err != nil {
		http.Error(w, "Gagal tambah process style: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

func (h *Handler) ProcessStyleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.ProcessStyleImportRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "body JSON tidak valid", http.StatusBadRequest)
		return
	}

	result, err := h.Repo.ImportProcessStyles(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ProcessStyleByID godoc
// @Summary Update atau delete proses style
// @Description Update atau hapus data proses style berdasarkan id
// @Tags Process Style
// @Accept json
// @Produce json
// @Param id path int true "ID process style"
// @Param body body models.ProcessStyleRequest false "Body process style untuk PUT"
// @Success 200 {object} models.ProcessStyleRecord
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /process-style/{id} [put]
// @Router /process-style/{id} [delete]
func (h *Handler) ProcessStyleByID(w http.ResponseWriter, r *http.Request) {
	id, err := getProcessStyleIDFromPath(r.URL.Path)
	if err != nil {
		http.Error(w, "id tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodPut:
		var input models.ProcessStyleRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Body JSON tidak valid: "+err.Error(), http.StatusBadRequest)
			return
		}

		input = normalizeProcessStyleInput(input)

		if input.ProcessName == "" {
			http.Error(w, "processName/proses wajib diisi", http.StatusBadRequest)
			return
		}

		if input.StyleName == "" {
			http.Error(w, "styleName/style wajib diisi", http.StatusBadRequest)
			return
		}

		data, err := h.Repo.UpdateProcessStyle(ctx, id, input)
		if err != nil {
			if errors.Is(err, repository.ErrProcessStyleNotFound) {
				http.Error(w, "Data process style tidak ditemukan", http.StatusNotFound)
				return
			}

			http.Error(w, "Gagal update process style: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, data)

	case http.MethodDelete:
		if err := h.Repo.DeleteProcessStyle(ctx, id); err != nil {
			if errors.Is(err, repository.ErrProcessStyleNotFound) {
				http.Error(w, "Data process style tidak ditemukan", http.StatusNotFound)
				return
			}

			http.Error(w, "Gagal hapus process style: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{
			"status":  "ok",
			"message": "Process style berhasil dihapus",
			"id":      id,
		})

	default:
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
	}
}

// ProcessStyleGccList godoc
// @Summary List Master IE dari cache lokal gcc_smv_process
// @Description Baca tabel lokal hasil sync GCC (cepat untuk UI & search)
// @Tags Process Style
// @Produce json
// @Param page query int false "Halaman"
// @Param limit query int false "Jumlah baris per halaman"
// @Param q query string false "Filter style/proses/buyer/item/mesin"
// @Success 200 {object} models.GccSmvListPageResponse
// @Failure 500 {string} string
// @Router /process-style/gcc [get]
func (h *Handler) ProcessStyleGccList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page")))
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	rows, total, err := h.Repo.ListGccSmvProcess(ctx, q, page, limit)
	if err != nil {
		http.Error(w, "Gagal ambil cache Master IE: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lastPage := total / limit
	if total%limit != 0 {
		lastPage++
	}
	if lastPage <= 0 {
		lastPage = 1
	}
	if page > lastPage {
		page = lastPage
	}
	from := 0
	to := 0
	if total > 0 && len(rows) > 0 {
		from = (page-1)*limit + 1
		to = from + len(rows) - 1
	}
	var nextPage *int
	var prevPage *int
	if page < lastPage {
		n := page + 1
		nextPage = &n
	}
	if page > 1 {
		p := page - 1
		prevPage = &p
	}

	msg := "Data Master IE dari tabel lokal gcc_smv_process"
	if q != "" {
		msg = "Hasil pencarian dari tabel lokal gcc_smv_process"
	}
	if total == 0 {
		msg = "Cache kosong. Klik Sinkron lokal dari GCC dulu."
	}

	writeJSON(w, models.GccSmvListPageResponse{
		Status:  "ok",
		Message: msg,
		Source:  "local-cache",
		Rows:    rows,
		Pagination: models.GccSmvListPagination{
			Total:       total,
			PerPage:     limit,
			CurrentPage: page,
			LastPage:    lastPage,
			From:        from,
			To:          to,
			NextPage:    nextPage,
			PrevPage:    prevPage,
		},
	})
}

// ProcessStyleSyncGcc godoc
// @Summary Sinkron Master IE dari GCC
// @Description Ambil semua detail-proses GCC, replace total tabel gcc_smv_process
// @Tags Process Style
// @Produce json
// @Success 200 {object} models.GccSmvSyncResponse
// @Failure 500 {string} string
// @Router /process-style/sync-gcc [post]
func (h *Handler) ProcessStyleSyncGcc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Sync penuh: banyak halaman GCC + upsert DB (bisa puluhan menit).
	ctx, cancel := contextWithTimeout(r, 45*time.Minute)
	defer cancel()

	rows, meta, headersFetched, err := utils.FetchAllGccSmvFlatRows(ctx)
	if err != nil {
		http.Error(w, "Gagal ambil data GCC: "+err.Error(), http.StatusBadGateway)
		return
	}

	stats, err := h.Repo.UpsertGccSmvProcessFromGcc(ctx, rows)
	if err != nil {
		http.Error(w, "Gagal upsert gcc_smv_process: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, models.GccSmvSyncResponse{
		Status: "ok",
		Message: fmt.Sprintf(
			"Sinkron incremental selesai. Baru: %d, update: %d, dihapus: %d.",
			stats.Inserted,
			stats.Updated,
			stats.Deleted,
		),
		PagesFetched:   meta.CurrentPage,
		HeadersFetched: headersFetched,
		TotalHeaders:   meta.Total,
		RowsPrepared:   len(rows),
		StylesInserted: stats.Inserted,
		StylesUpdated:  stats.Updated,
		StylesDeleted:  stats.Deleted,
		CtInserted:     stats.Inserted + stats.Updated,
		SkippedEmpty:   stats.SkippedEmpty,
	})
}
