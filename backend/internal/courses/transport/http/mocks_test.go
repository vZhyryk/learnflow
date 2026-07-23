package coursehttp_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	coursedomain "learnflow_backend/internal/courses/domain"
	coursehttp "learnflow_backend/internal/courses/transport/http"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

func newAuthMux(svc *mockService) *http.ServeMux {
	h := coursehttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	mux := http.NewServeMux()
	h.RegisterRoutes(mux, alice.Chain{}, alice.Chain{})
	return mux
}

// httpFixture wires a mockService-backed mux and a request builder for a single
// route, shared by every per-handler fixture in this package (loginFixture,
// registerFixture, ...). Embed it and add the handler-specific svcResult/svcErr
// fields on top.
type httpFixture struct {
	mux    *http.ServeMux
	newReq func(body string, urlParams map[string]string) *http.Request
}

func newHTTPFixture(svc *mockService, method, path string) *httpFixture {
	return &httpFixture{
		mux: newAuthMux(svc),
		newReq: func(body string, urlParams map[string]string) *http.Request {
			if len(urlParams) > 0 {
				path += "?"
			}

			for key, value := range urlParams {
				if value != "" && key != "" {
					path += fmt.Sprintf("%s=%s", key, url.QueryEscape(value))
				}
			}

			return httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
		},
	}
}

type mockService struct {
	archiveCourse   func(ctx context.Context, courseID string) error
	createCourse    func(ctx context.Context, req coursedomain.CreateCourseRequest) (string, error)
	deleteCourse    func(ctx context.Context, courseID string) error
	getCourseBySlug func(ctx context.Context, slug string) (*coursedomain.Course, error)
	publishCourse   func(ctx context.Context, courseID string) error
	updateCourse    func(ctx context.Context, req coursedomain.UpdateCourseRequest) error
	getAllCourses   func(ctx context.Context, getType coursedomain.CourseStatus, params pagination.Params) (courseList []*coursedomain.Course, err error)
}

func (m *mockService) ArchiveCourse(ctx context.Context, courseID string) error {
	if m.archiveCourse == nil {
		panic("mockService.archiveCourse not set")
	}
	return m.archiveCourse(ctx, courseID)
}
func (m *mockService) CreateCourse(ctx context.Context, req coursedomain.CreateCourseRequest) (string, error) {
	if m.createCourse == nil {
		panic("mockService.createCourse not set")
	}
	return m.createCourse(ctx, req)
}

func (m *mockService) DeleteCourse(ctx context.Context, courseID string) error {
	if m.deleteCourse == nil {
		panic("mockService.deleteCourse not set")
	}
	return m.deleteCourse(ctx, courseID)
}

func (m *mockService) GetCourseBySlug(ctx context.Context, slug string) (*coursedomain.Course, error) {
	if m.getCourseBySlug == nil {
		panic("mockService.getCourseBySlug not set")
	}
	return m.getCourseBySlug(ctx, slug)
}
func (m *mockService) PublishCourse(ctx context.Context, courseID string) error {
	if m.publishCourse == nil {
		panic("mockService.publishCourse not set")
	}
	return m.publishCourse(ctx, courseID)
}
func (m *mockService) UpdateCourse(ctx context.Context, req coursedomain.UpdateCourseRequest) error {
	if m.updateCourse == nil {
		panic("mockService.updateCourse not set")
	}
	return m.updateCourse(ctx, req)
}
func (m *mockService) GetAllCourses(ctx context.Context, getType coursedomain.CourseStatus, params pagination.Params) (courseList []*coursedomain.Course, err error) {
	if m.getAllCourses == nil {
		panic("mockService.getAllCourses not set")
	}
	return m.getAllCourses(ctx, getType, params)
}
