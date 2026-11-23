package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend_machine/models"
)

// GetLatestOperatorProcessByUUIDs mengambil proses login terakhir per UUID mesin
// (status apa pun), selama process_name tidak kosong.
func (r *Repository) GetLatestOperatorProcessByUUIDs(
	ctx context.Context,
	uuids []string,
) (map[string]models.MachineOperatorLastProcess, error) {
	out := make(map[string]models.MachineOperatorLastProcess)
	clean := make([]string, 0, len(uuids))
	seen := make(map[string]bool, len(uuids))
	for _, raw := range uuids {
		uuid := strings.TrimSpace(raw)
		if uuid == "" {
			continue
		}
		key := strings.ToLower(uuid)
		if seen[key] {
			continue
		}
		seen[key] = true
		clean = append(clean, uuid)
	}
	if len(clean) == 0 {
		return out, nil
	}

	const batchSize = 80
	for i := 0; i < len(clean); i += batchSize {
		end := i + batchSize
		if end > len(clean) {
			end = len(clean)
		}
		batch := clean[i:end]

		args := make([]any, 0, len(batch))
		params := make([]string, 0, len(batch))
		for j, uuid := range batch {
			name := fmt.Sprintf("u%d", j)
			params = append(params, "@"+name)
			args = append(args, sql.Named(name, strings.ToLower(uuid)))
		}

		query := fmt.Sprintf(`
WITH ranked AS (
	SELECT
		ISNULL(uuid, '') AS uuid,
		ISNULL(process_name, '') AS process_name,
		ISNULL(style_name, '') AS style_name,
		ISNULL(CONVERT(VARCHAR(19), login_time, 120), '') AS login_time,
		ROW_NUMBER() OVER (
			PARTITION BY LOWER(LTRIM(RTRIM(uuid)))
			ORDER BY login_time DESC, id DESC
		) AS rn
	FROM dbo.machine_operator_sessions
	WHERE LOWER(LTRIM(RTRIM(uuid))) IN (%s)
	  AND LTRIM(RTRIM(ISNULL(process_name, ''))) <> ''
)
SELECT uuid, process_name, style_name, login_time
FROM ranked
WHERE rn = 1;
`, strings.Join(params, ", "))

		rows, err := r.DB.QueryContext(ctx, query, args...)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var item models.MachineOperatorLastProcess
			if err := rows.Scan(&item.UUID, &item.ProcessName, &item.StyleName, &item.LoginTime); err != nil {
				rows.Close()
				return out, err
			}
			item.UUID = strings.TrimSpace(item.UUID)
			item.ProcessName = strings.TrimSpace(item.ProcessName)
			item.StyleName = strings.TrimSpace(item.StyleName)
			if item.UUID == "" || item.ProcessName == "" {
				continue
			}
			out[strings.ToLower(item.UUID)] = item
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return out, err
		}
	}

	pairs := make([][2]string, 0, len(out))
	keys := make([]string, 0, len(out))
	for key, item := range out {
		pairs = append(pairs, [2]string{item.StyleName, item.ProcessName})
		keys = append(keys, key)
	}
	lookup, err := r.LookupGccSmvProsesIDs(ctx, pairs)
	if err != nil {
		return out, err
	}
	for _, key := range keys {
		item := out[key]
		item.ProsesID = lookup[gccSmvProsesLookupKey(item.StyleName, item.ProcessName)]
		out[key] = item
	}

	return out, nil
}
