package utils

func StatusFromPct(pct float64) string {
	if pct >= 90 {
		return "GOOD"
	}

	if pct >= 80 {
		return "NORMAL"
	}

	return "BAD"
}
