package repository

import "strings"

func containsLikePattern(value string) string {
	escaped := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(strings.ToLower(value))
	return "%" + escaped + "%"
}
