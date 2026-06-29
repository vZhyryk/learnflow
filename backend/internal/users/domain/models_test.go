package usersdomain

import (
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
			So((&ChangeUserProfileRequest{FirstName: &firstName}).Validate(), ShouldEqual, ErrFirstNameInvalid)
		})
		Convey("100 runes is valid", func() {
			firstName := strings.Repeat("а", 100)
			So((&ChangeUserProfileRequest{FirstName: &firstName}).Validate(), ShouldBeNil)
		})
		Convey("101 runes is invalid", func() {
			firstName := strings.Repeat("а", 101)
			So((&ChangeUserProfileRequest{FirstName: &firstName}).Validate(), ShouldEqual, ErrFirstNameInvalid)
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
			So((&ChangeUserProfileRequest{LastName: &lastName}).Validate(), ShouldEqual, ErrLastNameTooLong)
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
			So((&ChangeUserProfileRequest{PhoneNumber: &phone}).Validate(), ShouldEqual, ErrPhoneNumberInvalid)
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
			So((&ChangeUserProfileRequest{Country: &country}).Validate(), ShouldEqual, ErrCountryInvalid)
		})
		Convey("3-char code is invalid", func() {
			country := "UKR"
			So((&ChangeUserProfileRequest{Country: &country}).Validate(), ShouldEqual, ErrCountryInvalid)
		})
		Convey("empty string is invalid", func() {
			country := ""
			So((&ChangeUserProfileRequest{Country: &country}).Validate(), ShouldEqual, ErrCountryInvalid)
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
			So((&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate(), ShouldEqual, ErrDateOfBirthInvalid)
		})
		Convey("future date is invalid", func() {
			dob := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
			So((&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate(), ShouldEqual, ErrDateOfBirthInvalid)
		})
		Convey("malformed date is invalid", func() {
			dob := "15-06-1990"
			So((&ChangeUserProfileRequest{DateOfBirth: &dob}).Validate(), ShouldEqual, ErrDateOfBirthInvalid)
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
			So((&ChangeUserProfileRequest{Gender: &gender}).Validate(), ShouldEqual, ErrGenderInvalid)
		})
	})
}

func TestValidateUILanguage(t *testing.T) {
	Convey("UILanguage validation", t, func() {
		Convey("nil is allowed", func() {
			So((&ChangeUserProfileRequest{}).Validate(), ShouldBeNil)
		})
		for _, valid := range []string{"uk", "pl", "en"} {
			v := valid
			Convey(v+" is valid", func() {
				So((&ChangeUserProfileRequest{UILanguage: &v}).Validate(), ShouldBeNil)
			})
		}
		Convey("unsupported code is invalid", func() {
			lang := "de"
			So((&ChangeUserProfileRequest{UILanguage: &lang}).Validate(), ShouldEqual, ErrUILanguageInvalid)
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
			So((&ChangeUserProfileRequest{AvatarURL: &url}).Validate(), ShouldEqual, ErrAvatarURLInvalid)
		})
		Convey("URL without host is invalid", func() {
			url := "https://"
			So((&ChangeUserProfileRequest{AvatarURL: &url}).Validate(), ShouldEqual, ErrAvatarURLInvalid)
		})
		Convey("plain string is invalid", func() {
			url := "not-a-url"
			So((&ChangeUserProfileRequest{AvatarURL: &url}).Validate(), ShouldEqual, ErrAvatarURLInvalid)
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
			So((&ChangeUserProfileRequest{Timezone: &tz}).Validate(), ShouldEqual, ErrTimezoneInvalid)
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
			So((&ChangeUserProfileRequest{Bio: &bio}).Validate(), ShouldEqual, ErrBioTooLong)
		})
	})
}

// --- Apply ---

func TestApply(t *testing.T) {
	Convey("Apply merges non-nil fields into UserProfile", t, func() {
		base := &UserProfile{
			UserID:    "user-123",
			FirstName: "John",
			LastName:  "Doe",
			Country:   "UA",
			Gender:    "male",
		}

		Convey("non-nil fields are updated", func() {
			firstName, country := "Jane", "PL"
			ChangeUserProfileRequest{FirstName: &firstName, Country: &country}.Apply(base)
			So(base.FirstName, ShouldEqual, "Jane")
			So(base.Country, ShouldEqual, "PL")
		})

		Convey("nil fields are left unchanged", func() {
			firstName := "Jane"
			ChangeUserProfileRequest{FirstName: &firstName}.Apply(base)
			So(base.LastName, ShouldEqual, "Doe")
			So(base.Gender, ShouldEqual, "male")
		})

		Convey("empty request changes nothing", func() {
			ChangeUserProfileRequest{}.Apply(base)
			So(base.FirstName, ShouldEqual, "John")
			So(base.LastName, ShouldEqual, "Doe")
		})
	})
}
