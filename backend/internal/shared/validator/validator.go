package validator

import (
	"net/url"
	"regexp"
	"time"
	"unicode/utf8"
)

// Gender values are duplicated here rather than imported from a domain package
// to keep shared/validator free of a domain dependency (avoids shared -> domain -> shared cycles).
const (
	maleGender           = "male"
	femaleGender         = "female"
	otherGender          = "other"
	preferNotToSayGender = "prefer_not_to_say"
)

// EmailRX is the compiled regular expression for validating email addresses.
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// dobMinDate is the earliest date_of_birth accepted, matching the
// user_profiles_dob_min DB constraint.
var dobMinDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

// MatchesEmail reports whether value is a valid email address.
func MatchesEmail(value string) bool {
	if len(value) > 254 {
		return false
	}
	return EmailRX.MatchString(value)
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

// IsValidPhoneNumber reports whether value is at most 20 bytes, matching the
// user_profiles.phone_number varchar(20) column width.
func IsValidPhoneNumber(value string) bool {
	return len(value) <= 20
}

// IsValidCountryCode reports whether value is a 2-character ISO 3166-1
// alpha-2 country code, matching the user_profiles_country_check DB constraint.
func IsValidCountryCode(value string) bool {
	return utf8.RuneCountInString(value) == 2
}

// IsValidGender reports whether value is one of the accepted gender values,
// matching the user_profiles_gender_check DB constraint.
func IsValidGender(value string) bool {
	switch value {
	case maleGender, femaleGender, otherGender, preferNotToSayGender:
		return true
	default:
		return false
	}
}

// IsValidDateOfBirth reports whether value is a "2006-01-02"-formatted date
// within [1900-01-01, now], matching the user_profiles_dob_not_future and
// user_profiles_dob_min DB constraints.
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

// IsValidAvatarURL reports whether value is a valid HTTPS URL.
func IsValidAvatarURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != ""
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
