package validator

import (
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// NormalizePassword applies Unicode NFC normalization so visually identical passwords
// with different byte representations (e.g. combining vs precomposed accents) hash the same.
func NormalizePassword(password string) string {
	return norm.NFC.String(password)
}

// Duplicated (not imported from a domain package) to avoid a shared -> domain -> shared cycle.
const (
	maleGender           = "male"
	femaleGender         = "female"
	otherGender          = "other"
	preferNotToSayGender = "prefer_not_to_say"
)

// EmailRX matches valid email addresses.
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// SlugRX matches URL-safe slugs.
var SlugRX = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// UUIDRX matches canonical hyphenated UUIDs (gen_random_uuid() format).
var UUIDRX = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID reports whether value is a canonical UUID.
func IsValidUUID(value string) bool {
	return UUIDRX.MatchString(value)
}

// dobMinDate matches the user_profiles_dob_min DB constraint.
var dobMinDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

// MatchesEmail reports whether value is a valid email address.
func MatchesEmail(value string) bool {
	if len(value) > 254 {
		return false
	}
	return EmailRX.MatchString(value)
}

// IsValidSlug reports whether value is a URL-safe slug.
func IsValidSlug(value string) bool {
	return SlugRX.MatchString(value)
}

// IsValidFirstName reports whether value is 1-100 runes.
func IsValidFirstName(value string) bool {
	n := utf8.RuneCountInString(value)
	return n > 0 && n <= 100
}

// IsValidLastName reports whether value is at most 100 runes.
func IsValidLastName(value string) bool {
	return utf8.RuneCountInString(value) <= 100
}

// IsValidPhoneNumber reports whether value is at most 20 bytes (varchar(20)).
func IsValidPhoneNumber(value string) bool {
	return len(value) <= 20
}

// IsValidCountryCode reports whether value is a 2-letter ISO 3166-1 code.
func IsValidCountryCode(value string) bool {
	return utf8.RuneCountInString(value) == 2
}

// IsValidGender reports whether value is an accepted gender value.
func IsValidGender(value string) bool {
	switch value {
	case maleGender, femaleGender, otherGender, preferNotToSayGender:
		return true
	default:
		return false
	}
}

// IsValidDateOfBirth reports whether value is a "2006-01-02" date within [1900-01-01, now].
func IsValidDateOfBirth(value string) bool {
	dob, err := time.Parse("2006-01-02", value)
	if err != nil {
		return false
	}
	return !dob.After(time.Now().UTC()) && !dob.Before(dobMinDate)
}

// IsValidUILanguage reports whether value is a supported UI language code.
func IsValidUILanguage(value string) bool {
	switch value {
	case "uk", "pl", "ru", "en":
		return true
	default:
		return false
	}
}

// maxURLLength is a storage bound, not an SSRF control — URLs aren't fetched server-side.
const maxURLLength = 2048

// IsValidHTTPSURL reports whether value is an HTTPS URL with a host, within maxURLLength.
func IsValidHTTPSURL(value string) bool {
	if len(value) > maxURLLength {
		return false
	}
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

// IsValidAvatarURL reports whether value is a valid HTTPS URL.
func IsValidAvatarURL(value string) bool {
	return IsValidHTTPSURL(value)
}

// IsValidTimezone reports whether value is a valid IANA timezone name.
func IsValidTimezone(value string) bool {
	_, err := time.LoadLocation(value)
	return err == nil
}

// IsValidBio reports whether value is at most 500 runes.
func IsValidBio(value string) bool {
	return utf8.RuneCountInString(value) <= 500
}

// IsValidContentTitle reports whether value is 1-300 runes after trimming.
func IsValidContentTitle(value string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(value))
	return n > 0 && n <= 300
}

// IsValidContentDescription reports whether value is at most 10000 runes.
func IsValidContentDescription(value string) bool {
	return utf8.RuneCountInString(value) <= 10000
}

// IsValidSeoTitle reports whether value is at most 70 runes (search result truncation limit).
func IsValidSeoTitle(value string) bool {
	return utf8.RuneCountInString(value) <= 70
}

// IsValidSeoDescription reports whether value is at most 160 runes (search result truncation limit).
func IsValidSeoDescription(value string) bool {
	return utf8.RuneCountInString(value) <= 160
}

// IsValidContentBody reports whether value is at most 10000 runes.
func IsValidContentBody(value string) bool {
	return utf8.RuneCountInString(value) <= 10000
}
