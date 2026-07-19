package handlers

import (
	"net/http"
	"strings"
	"time"
)

// EmployeeSearch godoc
// @Summary Search employee
// @Description Cari data karyawan berdasarkan NIK atau nama
// @Tags Employees
// @Produce json
// @Param q query string true "NIK atau nama karyawan"
// @Success 200 {array} models.Employee
// @Failure 400 {string} string
// @Failure 405 {string} string
// @Failure 500 {string} string
// @Router /employees/search [get]
func (h *Handler) EmployeeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query q wajib diisi", http.StatusBadRequest)
		return
	}

	data, err := h.Repo.SearchEmployees(ctx, q, 20)
	if err != nil {
		http.Error(w, "Gagal cari employee: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}
