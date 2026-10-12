package handlers

import (
	"context"
	"strings"
	"time"

	"backend_machine/models"
	"backend_machine/utils"
)

type aiProductivityLoader func(context.Context, string, string) (models.APIResponse, error)

// prepareAiChatDateContext replaces the client snapshot with requested database data.
func prepareAiChatDateContext(ctx context.Context, input *models.AiChatRequest, now time.Time, load aiProductivityLoader) error {
	date, mentioned, err := utils.ExtractAiQuestionDate(input.Question, now)
	if err != nil {
		return err
	}
	if !mentioned {
		return nil
	}
	utils.ResolveAiRankingDateFollowup(input, date)
	if !utils.AiQuestionNeedsMachineData(input.Question) {
		return nil
	}
	shift := strings.TrimSpace(input.Shift)
	// CURRENT follows the server clock; dated queries use all shifts by default.
	if shift == "" || utils.NormalizeShiftCode(shift) == utils.ShiftCurrent {
		shift = utils.ShiftALL
	}
	daily, err := load(ctx, date, shift)
	if err != nil {
		return err
	}
	machines := make([]models.AiInsightMachineInput, 0, len(daily.Rows))
	location := strings.ToUpper(strings.TrimSpace(input.Location))
	status := strings.ToUpper(strings.TrimSpace(input.Status))
	for _, row := range daily.Rows {
		if location != "" && location != "ALL" && !strings.Contains(strings.ToUpper(row.Location), location) {
			continue
		}
		if status != "" && status != "ALL" && !strings.EqualFold(row.Status, status) {
			continue
		}
		name := row.NickName
		if name == "" {
			name = row.MachineName
		}
		machines = append(machines, models.AiInsightMachineInput{
			UUID: row.UUID, MachineName: name, Location: row.Location, Status: row.Status,
			ProductivityPct: row.ProductivityPct, RuntimeSec: row.RuntimeSec,
			ProcSec: row.ProcSec, LossTimeSec: row.LossTimeSec,
		})
	}
	input.Date = date
	input.Shift = shift
	input.Machines = machines
	input.Insight = nil
	input.DataSource = "database"
	return nil
}
