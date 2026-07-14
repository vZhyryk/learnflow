package validator_test

import (
	"learnflow_backend/internal/shared/validator"
	"testing"

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
