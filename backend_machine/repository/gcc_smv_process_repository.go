package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend_machine/models"
)

const gccSmvInsertBatchSize = 80

type GccSmvUpsertStats struct {
	Inserted     int
	Updated      int
	Deleted      int
	SkippedEmpty int
}

func (r *Repository) EnsureGccSmvProcessSchema(ctx context.Context) error {
	query := `
IF OBJECT_ID(N'dbo.gcc_smv_process', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.gcc_smv_process (
		id BIGINT IDENTITY(1,1) NOT NULL CONSTRAINT PK_gcc_smv_process PRIMARY KEY,
		header_id BIGINT NULL,
		tanggal NVARCHAR(20) NULL,
		buyer NVARCHAR(200) NULL,
		style NVARCHAR(100) NOT NULL,
		item NVARCHAR(200) NULL,
		nama_proses NVARCHAR(500) NOT NULL,
		proses_id NVARCHAR(50) NULL,
		cycle_time DECIMAL(12, 4) NOT NULL CONSTRAINT DF_gcc_smv_process_cycle_time DEFAULT (0),
		smv_minute DECIMAL(18, 8) NULL,
		output_per_h DECIMAL(18, 8) NULL,
		mesin NVARCHAR(200) NULL,
		synced_at DATETIME2 NOT NULL CONSTRAINT DF_gcc_smv_process_synced_at DEFAULT SYSDATETIME()
	);

	CREATE INDEX IX_gcc_smv_process_style
		ON dbo.gcc_smv_process (style);

	CREATE UNIQUE INDEX UX_gcc_smv_process_style_proses
		ON dbo.gcc_smv_process (style, nama_proses);
END

IF OBJECT_ID(N'dbo.gcc_smv_process', N'U') IS NOT NULL
   AND COL_LENGTH('dbo.gcc_smv_process', 'proses_id') IS NULL
BEGIN
	ALTER TABLE dbo.gcc_smv_process ADD proses_id NVARCHAR(50) NULL;
END
`
	_, err := r.DB.ExecContext(ctx, query)
	return err
}

// UpsertGccSmvProcessFromGcc: insert baru + update yang berubah, lalu hapus yang tidak ada di GCC.
func (r *Repository) UpsertGccSmvProcessFromGcc(
	ctx context.Context,
	rows []models.GccSmvFlatRow,
) (GccSmvUpsertStats, error) {
	stats := GccSmvUpsertStats{}
	if err := r.EnsureGccSmvProcessSchema(ctx); err != nil {
		return stats, err
	}

	clean := make([]models.GccSmvFlatRow, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		style := cleanProcessStyleImportText(row.Style)
		proses := cleanProcessStyleImportText(row.NamaProses)
		if style == "" || proses == "" {
			stats.SkippedEmpty++
			continue
		}
		key := strings.ToLower(style) + "|" + strings.ToLower(proses)
		if seen[key] {
			stats.SkippedEmpty++
			continue
		}
		seen[key] = true

		ct := row.CycleTime
		if ct < 0 {
			ct = 0
		}
		clean = append(clean, models.GccSmvFlatRow{
			HeaderID:   row.HeaderID,
			Tanggal:    cleanProcessStyleImportText(row.Tanggal),
			Buyer:      cleanProcessStyleImportText(row.Buyer),
			Style:      style,
			Item:       cleanProcessStyleImportText(row.Item),
			NamaProses: proses,
			ProsesID:   cleanProcessStyleImportText(row.ProsesID),
			CycleTime:  ct,
			SmvMinute:  row.SmvMinute,
			OutputPerH: row.OutputPerH,
			Mesin:      cleanProcessStyleImportText(row.Mesin),
		})
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer tx.Rollback()

	syncAt := time.Now().UTC()

	for i := 0; i < len(clean); i += gccSmvInsertBatchSize {
		end := i + gccSmvInsertBatchSize
		if end > len(clean) {
			end = len(clean)
		}
		ins, upd, err := upsertGccSmvProcessBatch(ctx, tx, clean[i:end], syncAt)
		if err != nil {
			return stats, err
		}
		stats.Inserted += ins
		stats.Updated += upd
	}

	delResult, err := tx.ExecContext(
		ctx,
		`DELETE FROM dbo.gcc_smv_process WHERE synced_at < @sync_at;`,
		sql.Named("sync_at", syncAt),
	)
	if err != nil {
		return stats, fmt.Errorf("gagal hapus data GCC yang sudah tidak ada: %w", err)
	}
	if n, err := delResult.RowsAffected(); err == nil {
		stats.Deleted = int(n)
	}

	if err := tx.Commit(); err != nil {
		return stats, err
	}
	return stats, nil
}

func upsertGccSmvProcessBatch(
	ctx context.Context,
	tx *sql.Tx,
	batch []models.GccSmvFlatRow,
	syncAt time.Time,
) (inserted int, updated int, err error) {
	if len(batch) == 0 {
		return 0, 0, nil
	}

	var source strings.Builder
	args := make([]any, 0, len(batch)*11+1)
	args = append(args, sql.Named("sync_at", syncAt))

	for i, row := range batch {
		if i > 0 {
			source.WriteString(" UNION ALL ")
		}
		source.WriteString(fmt.Sprintf(
			`SELECT @h%d AS header_id, @t%d AS tanggal, @b%d AS buyer, @s%d AS style, @i%d AS item, @p%d AS nama_proses, @pid%d AS proses_id, @c%d AS cycle_time, @sm%d AS smv_minute, @o%d AS output_per_h, @m%d AS mesin`,
			i, i, i, i, i, i, i, i, i, i, i,
		))
		args = append(args,
			sql.Named(fmt.Sprintf("h%d", i), nullInt64(row.HeaderID)),
			sql.Named(fmt.Sprintf("t%d", i), nullString(row.Tanggal)),
			sql.Named(fmt.Sprintf("b%d", i), nullString(row.Buyer)),
			sql.Named(fmt.Sprintf("s%d", i), row.Style),
			sql.Named(fmt.Sprintf("i%d", i), nullString(row.Item)),
			sql.Named(fmt.Sprintf("p%d", i), row.NamaProses),
			sql.Named(fmt.Sprintf("pid%d", i), nullString(row.ProsesID)),
			sql.Named(fmt.Sprintf("c%d", i), row.CycleTime),
			sql.Named(fmt.Sprintf("sm%d", i), row.SmvMinute),
			sql.Named(fmt.Sprintf("o%d", i), row.OutputPerH),
			sql.Named(fmt.Sprintf("m%d", i), nullString(row.Mesin)),
		)
	}

	query := `
MERGE dbo.gcc_smv_process WITH (HOLDLOCK) AS target
USING (
` + source.String() + `
) AS source
ON LOWER(LTRIM(RTRIM(target.style))) = LOWER(LTRIM(RTRIM(source.style)))
 AND LOWER(LTRIM(RTRIM(target.nama_proses))) = LOWER(LTRIM(RTRIM(source.nama_proses)))
WHEN MATCHED THEN
	UPDATE SET
		header_id = source.header_id,
		tanggal = source.tanggal,
		buyer = source.buyer,
		style = source.style,
		item = source.item,
		nama_proses = source.nama_proses,
		proses_id = source.proses_id,
		cycle_time = source.cycle_time,
		smv_minute = source.smv_minute,
		output_per_h = source.output_per_h,
		mesin = source.mesin,
		synced_at = @sync_at
WHEN NOT MATCHED THEN
	INSERT (
		header_id, tanggal, buyer, style, item, nama_proses, proses_id,
		cycle_time, smv_minute, output_per_h, mesin, synced_at
	)
	VALUES (
		source.header_id, source.tanggal, source.buyer, source.style, source.item, source.nama_proses, source.proses_id,
		source.cycle_time, source.smv_minute, source.output_per_h, source.mesin, @sync_at
	)
OUTPUT $action;
`

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return 0, 0, fmt.Errorf("gagal upsert batch gcc_smv_process: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			return 0, 0, err
		}
		switch strings.ToUpper(strings.TrimSpace(action)) {
		case "INSERT":
			inserted++
		case "UPDATE":
			updated++
		}
	}
	return inserted, updated, rows.Err()
}

func nullString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func (r *Repository) ListGccSmvProcess(
	ctx context.Context,
	q string,
	page int,
	limit int,
) ([]models.GccSmvFlatRow, int, error) {
	if err := r.EnsureGccSmvProcessSchema(ctx); err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	offset := (page - 1) * limit

	countQuery := `
SELECT COUNT(1)
FROM dbo.gcc_smv_process
WHERE
	@q = ''
	OR style LIKE '%' + @q + '%'
	OR nama_proses LIKE '%' + @q + '%'
	OR ISNULL(proses_id, '') LIKE '%' + @q + '%'
	OR ISNULL(buyer, '') LIKE '%' + @q + '%'
	OR ISNULL(item, '') LIKE '%' + @q + '%'
	OR ISNULL(mesin, '') LIKE '%' + @q + '%';
`

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery, sql.Named("q", q)).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
SELECT
	ISNULL(header_id, 0),
	ISNULL(tanggal, ''),
	ISNULL(buyer, ''),
	style,
	ISNULL(item, ''),
	nama_proses,
	ISNULL(proses_id, ''),
	CAST(cycle_time AS FLOAT),
	CAST(ISNULL(smv_minute, 0) AS FLOAT),
	CAST(ISNULL(output_per_h, 0) AS FLOAT),
	ISNULL(mesin, '')
FROM dbo.gcc_smv_process
WHERE
	@q = ''
	OR style LIKE '%' + @q + '%'
	OR nama_proses LIKE '%' + @q + '%'
	OR ISNULL(proses_id, '') LIKE '%' + @q + '%'
	OR ISNULL(buyer, '') LIKE '%' + @q + '%'
	OR ISNULL(item, '') LIKE '%' + @q + '%'
	OR ISNULL(mesin, '') LIKE '%' + @q + '%'
ORDER BY
	TRY_CONVERT(date, NULLIF(LTRIM(RTRIM(tanggal)), '')) DESC,
	id DESC
OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY;
`

	rows, err := r.DB.QueryContext(
		ctx,
		listQuery,
		sql.Named("q", q),
		sql.Named("offset", offset),
		sql.Named("limit", limit),
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]models.GccSmvFlatRow, 0, limit)
	for rows.Next() {
		var item models.GccSmvFlatRow
		if err := rows.Scan(
			&item.HeaderID,
			&item.Tanggal,
			&item.Buyer,
			&item.Style,
			&item.Item,
			&item.NamaProses,
			&item.ProsesID,
			&item.CycleTime,
			&item.SmvMinute,
			&item.OutputPerH,
			&item.Mesin,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func (r *Repository) CountGccSmvProcess(ctx context.Context) (int, error) {
	if err := r.EnsureGccSmvProcessSchema(ctx); err != nil {
		return 0, err
	}
	var n int
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM dbo.gcc_smv_process;`).Scan(&n)
	return n, err
}

func gccSmvProsesLookupKey(style, namaProses string) string {
	return strings.ToLower(strings.TrimSpace(style)) + "|" + strings.ToLower(strings.TrimSpace(namaProses))
}

// LookupGccSmvProsesIDs mencocokkan style + nama_proses ke proses_id Master IE.
func (r *Repository) LookupGccSmvProsesIDs(
	ctx context.Context,
	pairs [][2]string,
) (map[string]string, error) {
	out := make(map[string]string)
	if len(pairs) == 0 {
		return out, nil
	}
	if err := r.EnsureGccSmvProcessSchema(ctx); err != nil {
		return out, err
	}

	uniq := make([][2]string, 0, len(pairs))
	seen := make(map[string]bool, len(pairs))
	for _, pair := range pairs {
		style := strings.TrimSpace(pair[0])
		proses := strings.TrimSpace(pair[1])
		if style == "" || proses == "" {
			continue
		}
		key := gccSmvProsesLookupKey(style, proses)
		if seen[key] {
			continue
		}
		seen[key] = true
		uniq = append(uniq, [2]string{style, proses})
	}
	if len(uniq) == 0 {
		return out, nil
	}

	const batchSize = 40
	for i := 0; i < len(uniq); i += batchSize {
		end := i + batchSize
		if end > len(uniq) {
			end = len(uniq)
		}
		batch := uniq[i:end]

		var where strings.Builder
		args := make([]any, 0, len(batch)*2)
		for j, pair := range batch {
			if j > 0 {
				where.WriteString(" OR ")
			}
			where.WriteString(fmt.Sprintf(
				`(LOWER(LTRIM(RTRIM(style))) = LOWER(LTRIM(RTRIM(@s%d))) AND LOWER(LTRIM(RTRIM(nama_proses))) = LOWER(LTRIM(RTRIM(@p%d))))`,
				j, j,
			))
			args = append(args,
				sql.Named(fmt.Sprintf("s%d", j), pair[0]),
				sql.Named(fmt.Sprintf("p%d", j), pair[1]),
			)
		}

		query := `
SELECT style, nama_proses, ISNULL(proses_id, '')
FROM dbo.gcc_smv_process
WHERE ` + where.String()

		rows, err := r.DB.QueryContext(ctx, query, args...)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var style, proses, prosesID string
			if err := rows.Scan(&style, &proses, &prosesID); err != nil {
				rows.Close()
				return out, err
			}
			prosesID = strings.TrimSpace(prosesID)
			if prosesID == "" {
				continue
			}
			out[gccSmvProsesLookupKey(style, proses)] = prosesID
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

func (r *Repository) EnrichOperatorSessionsProsesID(
	ctx context.Context,
	sessions []models.MachineOperatorSession,
) error {
	if len(sessions) == 0 {
		return nil
	}
	pairs := make([][2]string, 0, len(sessions))
	for _, s := range sessions {
		pairs = append(pairs, [2]string{s.StyleName, s.ProcessName})
	}
	lookup, err := r.LookupGccSmvProsesIDs(ctx, pairs)
	if err != nil {
		return err
	}
	for i := range sessions {
		key := gccSmvProsesLookupKey(sessions[i].StyleName, sessions[i].ProcessName)
		sessions[i].ProsesID = lookup[key]
	}
	return nil
}

func (r *Repository) EnrichOperatorReportProsesID(
	ctx context.Context,
	report []models.MachineOperatorReportItem,
) error {
	if len(report) == 0 {
		return nil
	}
	pairs := make([][2]string, 0, len(report))
	for _, item := range report {
		pairs = append(pairs, [2]string{item.StyleName, item.ProcessName})
	}
	lookup, err := r.LookupGccSmvProsesIDs(ctx, pairs)
	if err != nil {
		return err
	}
	for i := range report {
		key := gccSmvProsesLookupKey(report[i].StyleName, report[i].ProcessName)
		report[i].ProsesID = lookup[key]
	}
	return nil
}
