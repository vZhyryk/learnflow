package validator

import (
	"regexp"
)

// EmailRX is the compiled regular expression for validating email addresses.
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// MatchesEmail reports whether value is a valid email address.
func MatchesEmail(value string) bool {
	if len(value) > 254 {
		return false
	}
	return EmailRX.MatchString(value)
}
