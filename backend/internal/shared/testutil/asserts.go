package testutil

import (
	//nolint:staticcheck // project convention — dot-import Convey, same as every *_test.go file
	. "github.com/smartystreets/goconvey/convey"
)

// AssertUnexpectedDBError asserts err is non-nil and its message contains substr —
// the standard shape for a repository test's "unexpected error" branch.
func AssertUnexpectedDBError(err error, substr string) {
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, substr)
}
