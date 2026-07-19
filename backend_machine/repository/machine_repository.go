package repository

import (
	"context"
	"fmt"
	"strings"

	"backend_machine/models"
	"backend_machine/utils"
)

func (r *Repository) GetMachines(ctx context.Context) ([]models.Machine, error) {
	query := `
SELECT
    CONVERT(nvarchar(200), mi.UUID) AS UUID,
    ISNULL(CONVERT(nvarchar(200), mi.NickName), '') AS NickName,
    ISNULL(CONVERT(nvarchar(100), mi.LastLoginIP), '') AS LastLoginIP,
    LOWER('m' + CONVERT(nvarchar(200), mi.UUID)) AS TableName,
    ISNULL(CONVERT(nvarchar(200), mi.MacType), '') AS MacType,
    ISNULL(CONVERT(nvarchar(50), mi.MacState), '') AS MacState
FROM dbo.machineinfo mi
WHERE EXISTS (
    SELECT 1
    FROM sys.tables t
    WHERE LOWER(t.name) = LOWER('m' + CONVERT(nvarchar(200), mi.UUID))
)
ORDER BY mi.NickName, mi.UUID;
`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []models.Machine
	for rows.Next() {
		var m models.Machine
		if err := rows.Scan(&m.UUID, &m.NickName, &m.IP, &m.TableName, &m.MacType, &m.MacState); err != nil {
			return nil, err
		}

		m.UUID = strings.TrimSpace(m.UUID)
		m.NickName = utils.CleanDisplayText(m.NickName)
		m.IP = utils.CleanDisplayText(m.IP)
		m.TableName = strings.TrimSpace(m.TableName)
		m.MacType = utils.CleanDisplayText(m.MacType)
		m.MacState = utils.CleanDisplayText(m.MacState)

		if strings.TrimSpace(m.NickName) == "" {
			m.NickName = m.UUID
		}

		machines = append(machines, m)
	}

	return machines, rows.Err()
}

func (r *Repository) FindMachineByUUID(ctx context.Context, uuid string) (models.Machine, error) {
	machines, err := r.GetMachines(ctx)
	if err != nil {
		return models.Machine{}, err
	}

	for _, m := range machines {
		if strings.EqualFold(m.UUID, uuid) {
			return m, nil
		}
	}

	return models.Machine{}, fmt.Errorf("mesin UUID %s tidak ditemukan", uuid)
}
