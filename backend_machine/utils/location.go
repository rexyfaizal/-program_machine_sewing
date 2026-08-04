package utils

import (
	"regexp"
	"strings"
)

var gmLocationPattern = regexp.MustCompile(`(?i)\bGM\s*([0-9]+)\b`)

// LocationGroup menormalisasi teks lokasi menjadi grup area, contoh: "GM3 LINE 1" → "GM3".
func LocationGroup(location string) string {
	text := strings.TrimSpace(location)
	if text == "" {
		return "-"
	}

	text = strings.Join(strings.Fields(strings.ToUpper(text)), " ")

	if match := gmLocationPattern.FindStringSubmatch(text); len(match) == 2 {
		return "GM" + match[1]
	}

	return text
}

func IsGM3Location(location string) bool {
	return LocationGroup(location) == "GM3"
}

// ParseLocationParts memecah "GM3 LINE 1" / "GM3 - LINE 1" menjadi factory + line.
func ParseLocationParts(location string) (factory string, lineName string) {
	text := strings.TrimSpace(location)
	if text == "" {
		return "", ""
	}

	text = strings.Join(strings.Fields(strings.ToUpper(text)), " ")
	factory = LocationGroup(text)
	if factory == "-" {
		factory = ""
	}

	if factory == "" {
		return "", text
	}

	lineName = text
	prefixes := []string{
		factory + " - ",
		factory + " ",
		factory + "-",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(lineName, prefix) {
			lineName = strings.TrimSpace(strings.TrimPrefix(lineName, prefix))
			break
		}
	}

	if lineName == factory {
		lineName = ""
	}

	return factory, lineName
}

func LineShiftConfigKey(factory, lineName string) string {
	return strings.ToUpper(strings.TrimSpace(factory)) + "||" + strings.ToUpper(strings.TrimSpace(lineName))
}
