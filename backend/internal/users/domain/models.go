package usersdomain

import (
	"net/url"
	"time"
	"unicode/utf8"
)

// UserProfile holds a user's extended profile information.
type UserProfile struct {
	UserID      string
	FirstName   string
	LastName    string
	PhoneNumber string
	Country     string
	City        string
	DateOfBirth *string
	Gender      string
	UILanguage  string
	AvatarURL   string
	Timezone    string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ChangeUserProfileRequest carries optional fields for a profile update; nil fields are left unchanged.
type ChangeUserProfileRequest struct {
	UserID      *string
	FirstName   *string
	LastName    *string
	PhoneNumber *string
	Country     *string
	City        *string
	DateOfBirth *string
	Gender      *string
	UILanguage  *string
	AvatarURL   *string
	Timezone    *string
	Bio         *string
}

// Validate checks all non-nil fields against their domain constraints.
func (req *ChangeUserProfileRequest) Validate() error {
	checks := []func() error{
		req.validateFirstName,
		req.validateLastName,
		req.validatePhoneNumber,
		req.validateCountry,
		req.validateDateOfBirth,
		req.validateGender,
		req.validateUILanguage,
		req.validateAvatarURL,
		req.validateTimezone,
		req.validateBio,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateFirstName() error {
	if req.FirstName == nil {
		return nil
	}
	n := utf8.RuneCountInString(*req.FirstName)
	if n == 0 || n > 100 {
		return ErrFirstNameInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateLastName() error {
	if req.LastName != nil && utf8.RuneCountInString(*req.LastName) > 100 {
		return ErrLastNameTooLong
	}
	return nil
}

func (req *ChangeUserProfileRequest) validatePhoneNumber() error {
	if req.PhoneNumber != nil && len(*req.PhoneNumber) > 20 {
		return ErrPhoneNumberInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateCountry() error {
	if req.Country != nil && utf8.RuneCountInString(*req.Country) != 2 {
		return ErrCountryInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateDateOfBirth() error {
	if req.DateOfBirth == nil {
		return nil
	}
	dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
	if err != nil {
		return ErrDateOfBirthInvalid
	}
	minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if dob.After(time.Now().UTC()) || dob.Before(minDate) {
		return ErrDateOfBirthInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateGender() error {
	if req.Gender == nil {
		return nil
	}
	switch *req.Gender {
	case "male", "female", "other", "prefer_not_to_say":
		return nil
	default:
		return ErrGenderInvalid
	}
}

func (req *ChangeUserProfileRequest) validateUILanguage() error {
	if req.UILanguage == nil {
		return nil
	}
	switch *req.UILanguage {
	case "uk", "pl", "en":
		return nil
	default:
		return ErrUILanguageInvalid
	}
}

func (req *ChangeUserProfileRequest) validateAvatarURL() error {
	if req.AvatarURL == nil {
		return nil
	}
	u, err := url.Parse(*req.AvatarURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ErrAvatarURLInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateTimezone() error {
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			return ErrTimezoneInvalid
		}
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateBio() error {
	if req.Bio != nil && utf8.RuneCountInString(*req.Bio) > 500 {
		return ErrBioTooLong
	}
	return nil
}

// Apply merges non-nil request fields into the given UserProfile.
func (r ChangeUserProfileRequest) Apply(p *UserProfile) {
	applyString(&p.FirstName, r.FirstName)
	applyString(&p.LastName, r.LastName)
	applyString(&p.PhoneNumber, r.PhoneNumber)
	applyString(&p.Country, r.Country)
	applyString(&p.City, r.City)
	applyString(p.DateOfBirth, r.DateOfBirth)
	applyString(&p.Gender, r.Gender)
	applyString(&p.UILanguage, r.UILanguage)
	applyString(&p.AvatarURL, r.AvatarURL)
	applyString(&p.Timezone, r.Timezone)
	applyString(&p.Bio, r.Bio)
}

func applyString(dst, src *string) {
	if src != nil {
		*dst = *src
	}
}
