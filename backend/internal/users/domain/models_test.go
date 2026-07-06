package usersdomain

import (
	"errors"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// --- Validate ---

func TestValidateFirstName(t *testing.T) {
	Convey("FirstName validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("empty string is invalid", func() {
			firstName := ""
			err := (&ChangeUserProfileRequest{FirstName: &firstName}).Validate()
			So(errors.Is(err, ErrFirstNameInvalid), ShouldBeTrue)
		})
		Convey("100 runes is valid", func() {
			firstName := strings.Repeat("а", 100)
			So((&ChangeUserProfileRequest{FirstName: &firstName}).Validate(), ShouldBeNil)
		})
		Convey("101 runes is invalid", func() {
			firstName := strings.Repeat("а", 101)
			err := (&ChangeUserProfileRequest{FirstName: &firstName}).Validate()
			So(errors.Is(err, ErrFirstNameInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateLastName(t *testing.T) {
	Convey("LastName validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("empty string is allowed", func() {
			lastName := ""
			So((&ChangeUserProfileRequest{LastName: &lastName}).Validate(), ShouldBeNil)
		})
		Convey("100 runes is valid", func() {
			lastName := strings.Repeat("б", 100)
			So((&ChangeUserProfileRequest{LastName: &lastName}).Validate(), ShouldBeNil)
		})
		Convey("101 runes is invalid", func() {
			lastName := strings.Repeat("б", 101)
			err := (&ChangeUserProfileRequest{LastName: &lastName}).Validate()
			So(errors.Is(err, ErrLastNameTooLong), ShouldBeTrue)
		})
	})
}

func TestValidatePhoneNumber(t *testing.T) {
	Convey("PhoneNumber validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("20 chars is valid", func() {
			phone := strings.Repeat("1", 20)
			So((&ChangeUserProfileRequest{PhoneNumber: &phone}).Validate(), ShouldBeNil)
		})
		Convey("21 chars is invalid", func() {
			phone := strings.Repeat("1", 21)
			err := (&ChangeUserProfileRequest{PhoneNumber: &phone}).Validate()
			So(errors.Is(err, ErrPhoneNumberInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateCountry(t *testing.T) {
	Convey("Country validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("2-char code is valid", func() {
			country := "UA"
			So((&ChangeUserProfileRequest{Country: &country}).Validate(), ShouldBeNil)
		})
		Convey("1-char code is invalid", func() {
			country := "U"
			err := (&ChangeUserProfileRequest{Country: &country}).Validate()
			So(errors.Is(err, ErrCountryInvalid), ShouldBeTrue)
		})
		Convey("3-char code is invalid", func() {
			country := "UKR"
			err := (&ChangeUserProfileRequest{Country: &country}).Validate()
			So(errors.Is(err, ErrCountryInvalid), ShouldBeTrue)
		})
		Convey("empty string is invalid", func() {
			country := ""
			err := (&ChangeUserProfileRequest{Country: &country}).Validate()
			So(errors.Is(err, ErrCountryInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateDateOfBirth(t *testing.T) {
	Convey("DateOfBirth validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("valid past date is accepted", func() {
			dob := "1990-06-15"
			So((&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate(), ShouldBeNil)
		})
		Convey("1900-01-01 boundary is accepted", func() {
			dob := "1900-01-01"
			So((&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate(), ShouldBeNil)
		})
		Convey("date before 1900 is invalid", func() {
			dob := "1899-12-31"
			err := (&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate()
			So(errors.Is(err, ErrDateOfBirthInvalid), ShouldBeTrue)
		})
		Convey("future date is invalid", func() {
			dob := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
			err := (&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate()
			So(errors.Is(err, ErrDateOfBirthInvalid), ShouldBeTrue)
		})
		Convey("malformed date is invalid", func() {
			dob := "15-06-1990"
			err := (&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate()
			So(errors.Is(err, ErrDateOfBirthInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateGender(t *testing.T) {
	Convey("Gender validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		for _, valid := range []string{"male", "female", "other", "prefer_not_to_say"} {
			v := valid
			Convey(v+" is valid", func() {
				So((&ChangeUserProfileRequest{Gender: &v}).Validate(), ShouldBeNil)
			})
		}
		Convey("unknown value is invalid", func() {
			gender := "unknown"
			err := (&ChangeUserProfileRequest{Gender: &gender}).Validate()
			So(errors.Is(err, ErrGenderInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateUILanguage(t *testing.T) {
	Convey("UILanguage validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		for _, valid := range []string{"uk", "pl", "ru", "en"} {
			v := valid
			Convey(v+" is valid", func() {
				So((&ChangeUserProfileRequest{UILanguage: &v}).Validate(), ShouldBeNil)
			})
		}
		Convey("unsupported code is invalid", func() {
			lang := "de"
			err := (&ChangeUserProfileRequest{UILanguage: &lang}).Validate()
			So(errors.Is(err, ErrUILanguageInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateAvatarURL(t *testing.T) {
	Convey("AvatarURL validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("valid https URL is accepted", func() {
			url := "https://cdn.example.com/avatar.jpg"
			So((&ChangeUserProfileRequest{AvatarURL: &url}).Validate(), ShouldBeNil)
		})
		Convey("http URL is invalid", func() {
			url := "http://cdn.example.com/avatar.jpg"
			err := (&ChangeUserProfileRequest{AvatarURL: &url}).Validate()
			So(errors.Is(err, ErrAvatarURLInvalid), ShouldBeTrue)
		})
		Convey("URL without host is invalid", func() {
			url := "https://"
			err := (&ChangeUserProfileRequest{AvatarURL: &url}).Validate()
			So(errors.Is(err, ErrAvatarURLInvalid), ShouldBeTrue)
		})
		Convey("plain string is invalid", func() {
			url := "not-a-url"
			err := (&ChangeUserProfileRequest{AvatarURL: &url}).Validate()
			So(errors.Is(err, ErrAvatarURLInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateTimezone(t *testing.T) {
	Convey("Timezone validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("valid IANA timezone is accepted", func() {
			tz := "Europe/Kiev"
			So((&ChangeUserProfileRequest{Timezone: &tz}).Validate(), ShouldBeNil)
		})
		Convey("UTC is accepted", func() {
			tz := "UTC"
			So((&ChangeUserProfileRequest{Timezone: &tz}).Validate(), ShouldBeNil)
		})
		Convey("invalid timezone is rejected", func() {
			tz := "Mars/Olympus"
			err := (&ChangeUserProfileRequest{Timezone: &tz}).Validate()
			So(errors.Is(err, ErrTimezoneInvalid), ShouldBeTrue)
		})
	})
}

func TestValidateBio(t *testing.T) {
	Convey("Bio validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		Convey("500 runes is valid", func() {
			bio := strings.Repeat("я", 500)
			So((&ChangeUserProfileRequest{Bio: &bio}).Validate(), ShouldBeNil)
		})
		Convey("501 runes is invalid", func() {
			bio := strings.Repeat("я", 501)
			err := (&ChangeUserProfileRequest{Bio: &bio}).Validate()
			So(errors.Is(err, ErrBioTooLong), ShouldBeTrue)
		})
	})
}

// --- Apply ---

func TestApply(t *testing.T) {
	Convey("Apply merges non-nil fields into UserProfile", t, func() {
		baseFirstName, baseLastName, baseCountry, baseGender := "John", "Doe", "UA", "male"
		base := &UserProfile{
			UserID:    "user-123",
			FirstName: &baseFirstName,
			LastName:  &baseLastName,
			Country:   &baseCountry,
			Gender:    &baseGender,
		}

		Convey("non-nil fields are updated", func() {
			firstName, country := "Jane", "PL"
			ChangeUserProfileRequest{FirstName: &firstName, Country: &country}.Apply(base)
			So(*base.FirstName, ShouldEqual, "Jane")
			So(*base.Country, ShouldEqual, "PL")
		})

		Convey("nil fields are left unchanged", func() {
			firstName := "Jane"
			ChangeUserProfileRequest{FirstName: &firstName}.Apply(base)
			So(*base.LastName, ShouldEqual, "Doe")
			So(*base.Gender, ShouldEqual, "male")
		})

		Convey("empty request changes nothing", func() {
			ChangeUserProfileRequest{}.Apply(base)
			So(*base.FirstName, ShouldEqual, "John")
			So(*base.LastName, ShouldEqual, "Doe")
		})

		Convey("setting DateOfBirth for the first time populates a previously nil pointer", func() {
			dob := "1990-01-01"
			So(func() {
				ChangeUserProfileRequest{DateOfBirth: &dob}.Apply(base)
			}, ShouldNotPanic)
			So(*base.DateOfBirth, ShouldEqual, "1990-01-01")
		})
	})
}
