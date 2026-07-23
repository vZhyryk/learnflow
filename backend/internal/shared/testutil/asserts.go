package testutil

import (
	"github.com/smartystreets/goconvey/convey"
)

// AssertUnexpectedDBError asserts err is non-nil and its message contains substr —
// the standard shape for a repository test's "unexpected error" branch.
func AssertUnexpectedDBError(err error, substr string) {
	convey.So(err, convey.ShouldNotBeNil)
	convey.So(err.Error(), convey.ShouldContainSubstring, substr)
}
