package usersdomain

import (
	"learnflow_backend/internal/shared/validator"
	"time"
)

// UserProfile holds a user's extended profile information. Fields backed by
// nullable columns are *string: nil means the column is SQL NULL ("not
// provided"), distinct from an empty string. UILanguage stays a plain string
// because its column is NOT NULL DEFAULT 'uk'.
type UserProfile struct {
	UserID      string    `json:"user_id"`
	FirstName   *string   `json:"first_name"`
	LastName    *string   `json:"last_name"`
	PhoneNumber *string   `json:"phone_number"`
	Country     *string   `json:"country"`
	City        *string   `json:"city"`
	DateOfBirth *string   `json:"date_of_birth"`
	Gender      *string   `json:"gender"`
	UILanguage  string    `json:"ui_language"`
	AvatarURL   *string   `json:"avatar_url"`
	Timezone    *string   `json:"timezone"`
	Bio         *string   `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChangeUserProfileRequest carries optional fields for a profile update; nil fields are left unchanged.
type ChangeUserProfileRequest struct {
	UserID      *string `json:"-"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	PhoneNumber *string `json:"phone_number"`
	Country     *string `json:"country"`
	City        *string `json:"city"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
	UILanguage  *string `json:"ui_language"`
	AvatarURL   *string `json:"avatar_url"`
	Timezone    *string `json:"timezone"`
	Bio         *string `json:"bio"`
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
	if req.FirstName != nil && !validator.IsValidFirstName(*req.FirstName) {
		return ErrFirstNameInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateLastName() error {
	if req.LastName != nil && !validator.IsValidLastName(*req.LastName) {
		return ErrLastNameTooLong
	}
	return nil
}

func (req *ChangeUserProfileRequest) validatePhoneNumber() error {
	if req.PhoneNumber != nil && !validator.IsValidPhoneNumber(*req.PhoneNumber) {
		return ErrPhoneNumberInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateCountry() error {
	if req.Country != nil && !validator.IsValidCountryCode(*req.Country) {
		return ErrCountryInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateDateOfBirth() error {
	if req.DateOfBirth != nil && !validator.IsValidDateOfBirth(*req.DateOfBirth) {
		return ErrDateOfBirthInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateGender() error {
	if req.Gender != nil && !validator.IsValidGender(*req.Gender) {
		return ErrGenderInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateUILanguage() error {
	if req.UILanguage != nil && !validator.IsValidUILanguage(*req.UILanguage) {
		return ErrUILanguageInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateAvatarURL() error {
	if req.AvatarURL != nil && !validator.IsValidAvatarURL(*req.AvatarURL) {
		return ErrAvatarURLInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateTimezone() error {
	if req.Timezone != nil && !validator.IsValidTimezone(*req.Timezone) {
		return ErrTimezoneInvalid
	}
	return nil
}

func (req *ChangeUserProfileRequest) validateBio() error {
	if req.Bio != nil && !validator.IsValidBio(*req.Bio) {
		return ErrBioTooLong
	}
	return nil
}

// Apply merges non-nil request fields into the given UserProfile.
func (r ChangeUserProfileRequest) Apply(p *UserProfile) {
	r.applyIdentityFields(p)
	r.applyContactAndBioFields(p)
}

func (r ChangeUserProfileRequest) applyIdentityFields(p *UserProfile) {
	if r.FirstName != nil {
		p.FirstName = r.FirstName
	}
	if r.LastName != nil {
		p.LastName = r.LastName
	}
	if r.Country != nil {
		p.Country = r.Country
	}
	if r.City != nil {
		p.City = r.City
	}
	if r.DateOfBirth != nil {
		p.DateOfBirth = r.DateOfBirth
	}
	if r.Gender != nil {
		p.Gender = r.Gender
	}
}

func (r ChangeUserProfileRequest) applyContactAndBioFields(p *UserProfile) {
	if r.PhoneNumber != nil {
		p.PhoneNumber = r.PhoneNumber
	}
	if r.UILanguage != nil {
		p.UILanguage = *r.UILanguage
	}
	if r.AvatarURL != nil {
		p.AvatarURL = r.AvatarURL
	}
	if r.Timezone != nil {
		p.Timezone = r.Timezone
	}
	if r.Bio != nil {
		p.Bio = r.Bio
	}
}
