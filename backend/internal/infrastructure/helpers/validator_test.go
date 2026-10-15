package helpers

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
		Convey("when beforeValidateHook is provided, it runs before Validate", func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"value":"x"}`))
			req := &fakeValidatorReq{}

			hookCalled := false
			ok := DecodeAndValidate(w, r, testutil.NewTestLogger(), req, func() { hookCalled = true })
			So(ok, ShouldBeTrue)
			So(hookCalled, ShouldBeTrue)
		})
	})
}
