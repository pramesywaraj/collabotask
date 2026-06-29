package workspace

import "strings"

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
