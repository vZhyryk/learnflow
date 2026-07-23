package validator_test

import (
	"strings"
	"testing"
	"time"

	"learnflow_backend/internal/shared/validator"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMatchesEmail(t *testing.T) {
	Convey("MatchesEmail", t, func() {
		valid := []string{
			"user@example.com",
			"user.name+tag@sub.domain.org",
			"a@b.co",
			"user123@test-domain.io",
		}
		for _, v := range valid {
			Convey(v+" is valid", func() {
				So(validator.MatchesEmail(v), ShouldBeTrue)
			})
		}

		invalid := []string{
			"",
			"notanemail",
			"@nodomain.com",
			"user@",
			"user@domain",
			"user @domain.com",
			"user@domain.c",
		}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.MatchesEmail(v), ShouldBeFalse)
			})
		}
	})
}

func TestMatchesEmailLengthBoundary(t *testing.T) {
	Convey("MatchesEmail length boundary", t, func() {
		Convey("254 chars is valid (boundary)", func() {
			email := "a@" + strings.Repeat("a", 248) + ".com"
			So(len(email), ShouldEqual, 254)
			So(validator.MatchesEmail(email), ShouldBeTrue)
		})

		Convey("255 chars is invalid (boundary)", func() {
			email := "a@" + strings.Repeat("a", 249) + ".com"
			So(len(email), ShouldEqual, 255)
			So(validator.MatchesEmail(email), ShouldBeFalse)
		})
	})
}

func TestIsValidSlug(t *testing.T) {
	Convey("IsValidSlug", t, func() {
		valid := []string{
			"slug",
			"my-slug",
			"slug-123",
			"123",
			"a-b-c-d",
		}
		for _, v := range valid {
			Convey("'"+v+"' is valid", func() {
				So(validator.IsValidSlug(v), ShouldBeTrue)
			})
		}

		invalid := []string{
			"",
			"Slug",
			"my_slug",
			"-slug",
			"slug-",
			"my--slug",
			"my slug",
			"slug!",
		}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.IsValidSlug(v), ShouldBeFalse)
			})
		}
	})
}

func TestIsValidFirstName(t *testing.T) {
	Convey("IsValidFirstName", t, func() {
		Convey("empty is invalid", func() {
			So(validator.IsValidFirstName(""), ShouldBeFalse)
		})

		Convey("1 rune is valid (boundary)", func() {
			So(validator.IsValidFirstName("A"), ShouldBeTrue)
		})

		Convey("100 runes is valid (boundary)", func() {
			So(validator.IsValidFirstName(strings.Repeat("A", 100)), ShouldBeTrue)
		})

		Convey("101 runes is invalid (boundary)", func() {
			So(validator.IsValidFirstName(strings.Repeat("A", 101)), ShouldBeFalse)
		})

		Convey("multi-byte runes counted by rune, not byte", func() {
			So(validator.IsValidFirstName(strings.Repeat("Ж", 100)), ShouldBeTrue)
			So(validator.IsValidFirstName(strings.Repeat("Ж", 101)), ShouldBeFalse)
		})
	})
}

func TestIsValidLastName(t *testing.T) {
	Convey("IsValidLastName", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidLastName(""), ShouldBeTrue)
		})

		Convey("100 runes is valid (boundary)", func() {
			So(validator.IsValidLastName(strings.Repeat("A", 100)), ShouldBeTrue)
		})

		Convey("101 runes is invalid (boundary)", func() {
			So(validator.IsValidLastName(strings.Repeat("A", 101)), ShouldBeFalse)
		})

		Convey("multi-byte runes counted by rune, not byte", func() {
			So(validator.IsValidLastName(strings.Repeat("Ж", 100)), ShouldBeTrue)
			So(validator.IsValidLastName(strings.Repeat("Ж", 101)), ShouldBeFalse)
		})
	})
}

func TestIsValidPhoneNumber(t *testing.T) {
	Convey("IsValidPhoneNumber", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidPhoneNumber(""), ShouldBeTrue)
		})

		Convey("20 bytes is valid (boundary)", func() {
			So(validator.IsValidPhoneNumber(strings.Repeat("1", 20)), ShouldBeTrue)
		})

		Convey("21 bytes is invalid (boundary)", func() {
			So(validator.IsValidPhoneNumber(strings.Repeat("1", 21)), ShouldBeFalse)
		})

		Convey("typical formatted number is valid", func() {
			So(validator.IsValidPhoneNumber("+380501234567"), ShouldBeTrue)
		})
	})
}

func TestIsValidCountryCode(t *testing.T) {
	Convey("IsValidCountryCode", t, func() {
		Convey("2 letters is valid", func() {
			So(validator.IsValidCountryCode("US"), ShouldBeTrue)
		})

		Convey("empty is invalid", func() {
			So(validator.IsValidCountryCode(""), ShouldBeFalse)
		})

		Convey("1 char is invalid", func() {
			So(validator.IsValidCountryCode("U"), ShouldBeFalse)
		})

		Convey("3 chars is invalid", func() {
			So(validator.IsValidCountryCode("USA"), ShouldBeFalse)
		})

		Convey("counts runes, not bytes", func() {
			So(validator.IsValidCountryCode("ПЛ"), ShouldBeTrue)
		})
	})
}

func TestIsValidGender(t *testing.T) {
	Convey("IsValidGender", t, func() {
		valid := []string{"male", "female", "other", "prefer_not_to_say"}
		for _, v := range valid {
			Convey("'"+v+"' is valid", func() {
				So(validator.IsValidGender(v), ShouldBeTrue)
			})
		}

		invalid := []string{"", "Male", "unknown", "MALE", "other "}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.IsValidGender(v), ShouldBeFalse)
			})
		}
	})
}

func TestIsValidDateOfBirth(t *testing.T) {
	Convey("IsValidDateOfBirth", t, func() {
		Convey("malformed date is invalid", func() {
			So(validator.IsValidDateOfBirth("not-a-date"), ShouldBeFalse)
		})

		Convey("wrong format is invalid", func() {
			So(validator.IsValidDateOfBirth("01/02/2006"), ShouldBeFalse)
		})

		Convey("empty is invalid", func() {
			So(validator.IsValidDateOfBirth(""), ShouldBeFalse)
		})

		Convey("dobMinDate boundary (1900-01-01) is valid", func() {
			So(validator.IsValidDateOfBirth("1900-01-01"), ShouldBeTrue)
		})

		Convey("day before dobMinDate (1899-12-31) is invalid", func() {
			So(validator.IsValidDateOfBirth("1899-12-31"), ShouldBeFalse)
		})

		Convey("today (UTC) is valid boundary", func() {
			today := time.Now().UTC().Format("2006-01-02")
			So(validator.IsValidDateOfBirth(today), ShouldBeTrue)
		})

		Convey("a future date is invalid", func() {
			future := time.Now().UTC().AddDate(1, 0, 0).Format("2006-01-02")
			So(validator.IsValidDateOfBirth(future), ShouldBeFalse)
		})

		Convey("a normal past date is valid", func() {
			So(validator.IsValidDateOfBirth("1990-06-15"), ShouldBeTrue)
		})
	})
}

func TestIsValidUILanguage(t *testing.T) {
	Convey("IsValidUILanguage", t, func() {
		valid := []string{"uk", "pl", "ru", "en"}
		for _, v := range valid {
			Convey("'"+v+"' is valid", func() {
				So(validator.IsValidUILanguage(v), ShouldBeTrue)
			})
		}

		invalid := []string{"", "UK", "de", "fr", "english"}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.IsValidUILanguage(v), ShouldBeFalse)
			})
		}
	})
}

func TestIsValidHTTPSURL(t *testing.T) {
	Convey("IsValidHTTPSURL", t, func() {
		valid := []string{
			"https://example.com",
			"https://example.com/path",
			"https://sub.example.com:8443/path?query=1",
		}
		for _, v := range valid {
			Convey("'"+v+"' is valid", func() {
				So(validator.IsValidHTTPSURL(v), ShouldBeTrue)
			})
		}

		invalid := []string{
			"",
			"http://example.com",
			"ftp://example.com",
			"example.com",
			"https://",
			"not a url",
		}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.IsValidHTTPSURL(v), ShouldBeFalse)
			})
		}
	})
}

func TestIsValidAvatarURL(t *testing.T) {
	Convey("IsValidAvatarURL", t, func() {
		Convey("delegates to IsValidHTTPSURL — valid https URL", func() {
			So(validator.IsValidAvatarURL("https://example.com/avatar.png"), ShouldBeTrue)
		})

		Convey("delegates to IsValidHTTPSURL — non-https is invalid", func() {
			So(validator.IsValidAvatarURL("http://example.com/avatar.png"), ShouldBeFalse)
		})

		Convey("empty is invalid", func() {
			So(validator.IsValidAvatarURL(""), ShouldBeFalse)
		})
	})
}

func TestIsValidTimezone(t *testing.T) {
	Convey("IsValidTimezone", t, func() {
		valid := []string{"UTC", "Europe/Warsaw", "Europe/Kyiv", "America/New_York"}
		for _, v := range valid {
			Convey("'"+v+"' is valid", func() {
				So(validator.IsValidTimezone(v), ShouldBeTrue)
			})
		}

		Convey("empty string resolves to UTC via time.LoadLocation and is valid", func() {
			So(validator.IsValidTimezone(""), ShouldBeTrue)
		})

		invalid := []string{"Not/A_Timezone", "GMT+3"}
		for _, v := range invalid {
			Convey("'"+v+"' is invalid", func() {
				So(validator.IsValidTimezone(v), ShouldBeFalse)
			})
		}
	})
}

func TestIsValidBio(t *testing.T) {
	Convey("IsValidBio", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidBio(""), ShouldBeTrue)
		})

		Convey("500 runes is valid (boundary)", func() {
			So(validator.IsValidBio(strings.Repeat("A", 500)), ShouldBeTrue)
		})

		Convey("501 runes is invalid (boundary)", func() {
			So(validator.IsValidBio(strings.Repeat("A", 501)), ShouldBeFalse)
		})

		Convey("multi-byte runes counted by rune, not byte", func() {
			So(validator.IsValidBio(strings.Repeat("Ж", 500)), ShouldBeTrue)
			So(validator.IsValidBio(strings.Repeat("Ж", 501)), ShouldBeFalse)
		})
	})
}

func TestIsValidHTTPSURLLengthBoundary(t *testing.T) {
	Convey("IsValidHTTPSURL length boundary", t, func() {
		Convey("2048 bytes is valid (boundary)", func() {
			url := "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/"))
			So(len(url), ShouldEqual, 2048)
			So(validator.IsValidHTTPSURL(url), ShouldBeTrue)
		})

		Convey("2049 bytes is invalid (boundary)", func() {
			url := "https://example.com/" + strings.Repeat("a", 2049-len("https://example.com/"))
			So(len(url), ShouldEqual, 2049)
			So(validator.IsValidHTTPSURL(url), ShouldBeFalse)
		})
	})
}

func TestIsValidContentTitle(t *testing.T) {
	Convey("IsValidContentTitle", t, func() {
		Convey("empty is invalid", func() {
			So(validator.IsValidContentTitle(""), ShouldBeFalse)
		})

		Convey("whitespace-only is invalid", func() {
			So(validator.IsValidContentTitle("   "), ShouldBeFalse)
		})

		Convey("1 rune is valid (boundary)", func() {
			So(validator.IsValidContentTitle("A"), ShouldBeTrue)
		})

		Convey("300 runes is valid (boundary)", func() {
			So(validator.IsValidContentTitle(strings.Repeat("A", 300)), ShouldBeTrue)
		})

		Convey("301 runes is invalid (boundary)", func() {
			So(validator.IsValidContentTitle(strings.Repeat("A", 301)), ShouldBeFalse)
		})

		Convey("leading/trailing whitespace is trimmed before counting", func() {
			So(validator.IsValidContentTitle("  "+strings.Repeat("A", 300)+"  "), ShouldBeTrue)
		})
	})
}

func TestIsValidContentDescription(t *testing.T) {
	Convey("IsValidContentDescription", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidContentDescription(""), ShouldBeTrue)
		})

		Convey("10000 runes is valid (boundary)", func() {
			So(validator.IsValidContentDescription(strings.Repeat("A", 10000)), ShouldBeTrue)
		})

		Convey("10001 runes is invalid (boundary)", func() {
			So(validator.IsValidContentDescription(strings.Repeat("A", 10001)), ShouldBeFalse)
		})
	})
}

func TestIsValidSeoTitle(t *testing.T) {
	Convey("IsValidSeoTitle", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidSeoTitle(""), ShouldBeTrue)
		})

		Convey("70 runes is valid (boundary)", func() {
			So(validator.IsValidSeoTitle(strings.Repeat("A", 70)), ShouldBeTrue)
		})

		Convey("71 runes is invalid (boundary)", func() {
			So(validator.IsValidSeoTitle(strings.Repeat("A", 71)), ShouldBeFalse)
		})
	})
}

func TestIsValidSeoDescription(t *testing.T) {
	Convey("IsValidSeoDescription", t, func() {
		Convey("empty is valid", func() {
			So(validator.IsValidSeoDescription(""), ShouldBeTrue)
		})

		Convey("160 runes is valid (boundary)", func() {
			So(validator.IsValidSeoDescription(strings.Repeat("A", 160)), ShouldBeTrue)
		})

		Convey("161 runes is invalid (boundary)", func() {
			So(validator.IsValidSeoDescription(strings.Repeat("A", 161)), ShouldBeFalse)
		})
	})
}
