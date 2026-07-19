package utils

import "strings"

func CleanDisplayText(s string) string {
	// Bersihkan text yang rusak karena encoding Unicode/China menjadi tanda tanya.
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return s
	}

	// Kalau hasil baca dari SQL Server menjadi ????, jangan tampilkan.
	if strings.Contains(s, "?") {
		return ""
	}

	// Jangan tampilkan karakter non-ASCII seperti China/Jepang/Korea/dll.
	for _, r := range s {
		if r > 127 {
			return ""
		}
	}

	return s
}
