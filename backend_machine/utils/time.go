package utils

import "fmt"

func FormatDuration(sec int64) string {
	if sec < 0 {
		sec = 0
	}

	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
}
