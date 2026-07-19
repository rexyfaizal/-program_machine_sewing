package utils

import "strings"

func AbnormalReason(procCounts, endStitch, fileStitches int64) string {
	var reasons []string

	if procCounts == 0 {
		reasons = append(reasons, "ProcCounts = 0")
	}

	if fileStitches > 0 && endStitch < fileStitches {
		reasons = append(reasons, "EndStitch < FileStitches")
	}

	if len(reasons) == 0 {
		return "-"
	}

	return strings.Join(reasons, " | ")
}
