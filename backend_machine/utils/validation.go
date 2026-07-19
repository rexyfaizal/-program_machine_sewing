package utils

import "regexp"

func SafeTableName(name string) bool {
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, name)
	return ok
}
