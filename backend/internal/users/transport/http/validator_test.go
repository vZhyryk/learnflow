package usershttp

// This file is intentionally in the internal (non-_test-suffixed) package.
// Every production caller in this module invokes decodeAndValidate with a
// nil beforeValidateHook (see change_profile.go), so the "hook != nil" branch
// is unreachable through any real HTTP route and cannot be exercised via the
// exported mux — it can only be hit by calling decodeAndValidate directly.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeValidatorReq struct {
	Value string `json:"value"`
}

func (f *fakeValidatorReq) Validate() error { return nil }

func TestDecodeAndValidateBeforeHook(t *testing.T) {
	Convey("decodeAndValidate", t, func() {
		h := NewHTTPHandler(nil, testutil.NewTestLogger())

		Convey("when beforeValidateHook is provided, it runs before Validate", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"value":"x"}`))
			req := &fakeValidatorReq{}

			hookCalled := false
			ok := h.decodeAndValidate(w, r, req, func() { hookCalled = true })
			So(ok, ShouldBeTrue)
			So(hookCalled, ShouldBeTrue)
		})
	})
}
