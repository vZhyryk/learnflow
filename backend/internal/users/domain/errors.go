package usersdomain

import (
	"errors"
)

var (
	// ErrUserNotFound is returned when no user matches the lookup criteria.
	ErrUserNotFound = errors.New("user not found")
	// ErrFirstNameInvalid is returned when first name is empty or exceeds 100 characters.
	ErrFirstNameInvalid = errors.New("first name must be between 1 and 100 characters")
	// ErrLastNameTooLong is returned when last name exceeds 100 characters.
	ErrLastNameTooLong = errors.New("last name must not exceed 100 characters")
	// ErrPhoneNumberInvalid is returned when phone number exceeds 20 characters.
	ErrPhoneNumberInvalid = errors.New("phone number must not exceed 20 characters")
	// ErrCountryInvalid is returned when country is not a valid ISO 3166-1 alpha-2 code.
	ErrCountryInvalid = errors.New("country must be a valid ISO 3166-1 alpha-2 code (exactly 2 characters)")
	// ErrBioTooLong is returned when bio exceeds 500 characters.
	ErrBioTooLong = errors.New("bio must not exceed 500 characters")
	// ErrAvatarURLInvalid is returned when avatar URL is not a valid HTTPS URL.
	ErrAvatarURLInvalid = errors.New("avatar URL must be a valid HTTPS URL")
	// ErrGenderInvalid is returned when gender is not one of the allowed values.
	ErrGenderInvalid = errors.New("gender must be one of: male, female, other, prefer_not_to_say")
	// ErrUILanguageInvalid is returned when UI language is not a supported language code.
	ErrUILanguageInvalid = errors.New("UI language must be a supported language code")
	// ErrTimezoneInvalid is returned when timezone is not a valid IANA timezone.
	ErrTimezoneInvalid = errors.New("timezone must be a valid IANA timezone")
	// ErrDateOfBirthInvalid is returned when date of birth is malformed, in the future, or before 1900.
	ErrDateOfBirthInvalid = errors.New("date of birth must be a valid YYYY-MM-DD date, not in the future, not before 1900-01-01")
)
