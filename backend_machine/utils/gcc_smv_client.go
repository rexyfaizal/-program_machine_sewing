package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"backend_machine/models"
)

type GccSmvMachineProcess struct {
	ProcessID string `json:"process_id"`
	Machine   string `json:"machine"`
	Category  string `json:"category"`
}

type GccSmvDetailItem struct {
	NamaProses      string                 `json:"nama_proses"`
	CycleTime       float64                `json:"cycle_time"`
	SmvMinute       float64                `json:"smv_minute"`
	PrdOnCapacity   float64                `json:"prd_on_capacity"`
	Mesin           *string                `json:"mesin"`
	MachineProcess  *GccSmvMachineProcess  `json:"machine_process"`
}

type GccSmvHeaderItem struct {
	HeaderID int64              `json:"header_id"`
	Date     string             `json:"date"`
	Style    string             `json:"style"`
	Item     string             `json:"item"`
	Buyer    string             `json:"buyer"`
	Detail   []GccSmvDetailItem `json:"detail"`
}

type GccSmvPagination struct {
	Total       int  `json:"total"`
	PerPage     int  `json:"per_page"`
	CurrentPage int  `json:"current_page"`
	LastPage    int  `json:"last_page"`
	From        int  `json:"from"`
	To          int  `json:"to"`
	NextPage    *int `json:"next_page"`
	PrevPage    *int `json:"prev_page"`
}

type GccSmvListResponse struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Data       []GccSmvHeaderItem `json:"data"`
	Pagination GccSmvPagination   `json:"pagination"`
}

func gccBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("GCC_BASE_URL")), "/")
}

func gccAPIKey() string {
	return strings.TrimSpace(os.Getenv("GCC_API_KEY"))
}

func gccPageLimit() int {
	raw := strings.TrimSpace(os.Getenv("GCC_SMV_PAGE_LIMIT"))
	if raw == "" {
		// Default lebih kecil agar body JSON per halaman tidak terlalu berat.
		return 50
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 50
	}
	if n > 200 {
		return 200
	}
	return n
}

func gccHTTPTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("GCC_SMV_HTTP_TIMEOUT_SEC"))
	if raw == "" {
		return 5 * time.Minute
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 30 {
		return 5 * time.Minute
	}
	if n > 900 {
		return 15 * time.Minute
	}
	return time.Duration(n) * time.Second
}

func cleanGccText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.Join(strings.Fields(value), " ")
}

func fetchGccSmvPage(ctx context.Context, page, limit int) (GccSmvListResponse, error) {
	base := gccBaseURL()
	key := gccAPIKey()
	if base == "" {
		return GccSmvListResponse{}, fmt.Errorf("GCC_BASE_URL belum di-set di .env")
	}
	if key == "" {
		return GccSmvListResponse{}, fmt.Errorf("GCC_API_KEY belum di-set di .env")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = gccPageLimit()
	}
	if limit > 200 {
		limit = 200
	}

	endpoint, err := url.Parse(base + "/api/prd/smv/detail-proses")
	if err != nil {
		return GccSmvListResponse{}, err
	}
	q := endpoint.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return GccSmvListResponse{}, err
	}
	req.Header.Set("GCC-API-KEY", key)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: gccHTTPTimeout()}
	res, err := client.Do(req)
	if err != nil {
		return GccSmvListResponse{}, fmt.Errorf("GCC page %d gagal: %w", page, err)
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(res.Body, 64<<20))
	if readErr != nil {
		return GccSmvListResponse{}, fmt.Errorf("GCC page %d baca body gagal: %w", page, readErr)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return GccSmvListResponse{}, fmt.Errorf(
			"GCC page %d HTTP %d: %s",
			page,
			res.StatusCode,
			truncateGccErr(string(body), 240),
		)
	}

	var parsed GccSmvListResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return GccSmvListResponse{}, fmt.Errorf("GCC page %d JSON invalid: %w", page, err)
	}
	if !parsed.Success && len(parsed.Data) == 0 {
		msg := strings.TrimSpace(parsed.Message)
		if msg == "" {
			msg = "response success=false"
		}
		return GccSmvListResponse{}, fmt.Errorf("GCC page %d: %s", page, msg)
	}
	return parsed, nil
}

func mesinText(value *string) string {
	if value == nil {
		return ""
	}
	return cleanGccText(*value)
}

func FlattenGccSmvHeaders(headers []GccSmvHeaderItem) []models.GccSmvFlatRow {
	out := make([]models.GccSmvFlatRow, 0, 128)
	for _, header := range headers {
		style := cleanGccText(header.Style)
		buyer := cleanGccText(header.Buyer)
		item := cleanGccText(header.Item)
		tanggal := cleanGccText(header.Date)
		for _, detail := range header.Detail {
			proses := cleanGccText(detail.NamaProses)
			if style == "" || proses == "" {
				continue
			}
			ct := detail.CycleTime
			if ct < 0 {
				ct = 0
			}
			prosesID := ""
			if detail.MachineProcess != nil {
				prosesID = cleanGccText(detail.MachineProcess.ProcessID)
			}
			out = append(out, models.GccSmvFlatRow{
				HeaderID:   header.HeaderID,
				Tanggal:    tanggal,
				Buyer:      buyer,
				Style:      style,
				Item:       item,
				NamaProses: proses,
				ProsesID:   prosesID,
				CycleTime:  ct,
				SmvMinute:  detail.SmvMinute,
				OutputPerH: detail.PrdOnCapacity,
				Mesin:      mesinText(detail.Mesin),
			})
		}
	}
	return out
}

func FilterGccSmvFlatRows(rows []models.GccSmvFlatRow, keyword string) []models.GccSmvFlatRow {
	key := strings.ToLower(strings.TrimSpace(keyword))
	if key == "" {
		return rows
	}
	out := make([]models.GccSmvFlatRow, 0, len(rows))
	for _, row := range rows {
		blob := strings.ToLower(strings.Join([]string{
			row.Tanggal,
			row.Buyer,
			row.Style,
			row.Item,
			row.NamaProses,
			row.ProsesID,
			row.Mesin,
		}, " "))
		if strings.Contains(blob, key) {
			out = append(out, row)
		}
	}
	return out
}

// FetchGccSmvFlatPage ambil 1 halaman header GCC lalu flatten ke baris proses (untuk UI Master IE).
func FetchGccSmvFlatPage(ctx context.Context, page, limit int, keyword string) ([]models.GccSmvFlatRow, GccSmvPagination, error) {
	parsed, err := fetchGccSmvPage(ctx, page, limit)
	if err != nil {
		return nil, GccSmvPagination{}, err
	}
	rows := FilterGccSmvFlatRows(FlattenGccSmvHeaders(parsed.Data), keyword)
	return rows, parsed.Pagination, nil
}

// FetchAllGccSmvFlatRows menarik semua halaman GCC lalu flatten kolom lengkap.
// Jika style+proses duplikat, baris pertama (lebih baru) yang menang.
func FetchAllGccSmvFlatRows(ctx context.Context) ([]models.GccSmvFlatRow, GccSmvPagination, int, error) {
	limit := gccPageLimit()
	seen := make(map[string]bool)
	out := make([]models.GccSmvFlatRow, 0, 4096)
	pagesFetched := 0
	headersFetched := 0
	var lastMeta GccSmvPagination

	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return nil, lastMeta, pagesFetched, err
		}

		parsed, err := fetchGccSmvPage(ctx, page, limit)
		if err != nil {
			return nil, lastMeta, pagesFetched, err
		}

		pagesFetched++
		lastMeta = parsed.Pagination
		headersFetched += len(parsed.Data)

		for _, row := range FlattenGccSmvHeaders(parsed.Data) {
			keyPair := strings.ToLower(row.Style) + "|" + strings.ToLower(row.NamaProses)
			if seen[keyPair] {
				continue
			}
			seen[keyPair] = true
			out = append(out, row)
		}

		lastPage := parsed.Pagination.LastPage
		if lastPage <= 0 {
			if len(parsed.Data) < limit {
				break
			}
			page++
			continue
		}
		if page >= lastPage {
			break
		}
		page++
	}

	lastMeta.CurrentPage = pagesFetched
	return out, lastMeta, headersFetched, nil
}

func truncateGccErr(text string, max int) string {
	text = strings.TrimSpace(text)
	if max <= 0 || len(text) <= max {
		return text
	}
	return text[:max] + "..."
}
