// Package stringsx provides small string utility functions missing from the stdlib.
package stringsx

import (
	"strings"
)

// ToPascalFromSeparated converts a snake_case or kebab-case string to PascalCase.
func ToPascalFromSeparated(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})

	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}

	return strings.Join(parts, "")
}
