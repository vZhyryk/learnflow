package contenthttp_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	contentdomain "learnflow_backend/internal/content/domain"
	contenthttp "learnflow_backend/internal/content/transport/http"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/justinas/alice"
)

type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody
var withUser = testutil.WithUser

func newAuthMux(svc *mockService) *http.ServeMux {
	h := contenthttp.NewHTTPHandler(svc, testutil.NewTestLogger())
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
	archiveContentItem   func(ctx context.Context, contentItemID string) error
	createContentItem    func(ctx context.Context, req contentdomain.CreateContentItemRequest) (string, error)
	deleteContentItem    func(ctx context.Context, contentItemID string) error
	getContentItemBySlug func(ctx context.Context, slug string) (*contentdomain.ContentItem, error)
	publishContentItem   func(ctx context.Context, contentItemID string) error
	updateContentItem    func(ctx context.Context, req contentdomain.UpdateContentItemRequest) error
	getAllContentItems   func(ctx context.Context, getType contentdomain.ContentItemStatus, params pagination.Params) (contentItemList []*contentdomain.ContentItem, err error)
}

func (m *mockService) ArchiveContentItem(ctx context.Context, contentItemID string) error {
	if m.archiveContentItem == nil {
		panic("mockService.ArchiveContentItem not set")
	}
	return m.archiveContentItem(ctx, contentItemID)
}
func (m *mockService) CreateContentItem(ctx context.Context, req contentdomain.CreateContentItemRequest) (string, error) {
	if m.createContentItem == nil {
		panic("mockService.CreateContentItem not set")
	}
	return m.createContentItem(ctx, req)
}

func (m *mockService) DeleteContentItem(ctx context.Context, contentItemID string) error {
	if m.deleteContentItem == nil {
		panic("mockService.DeleteContentItem not set")
	}
	return m.deleteContentItem(ctx, contentItemID)
}

func (m *mockService) GetContentItemBySlug(ctx context.Context, slug string) (*contentdomain.ContentItem, error) {
	if m.getContentItemBySlug == nil {
		panic("mockService.GetContentItemBySlug not set")
	}
	return m.getContentItemBySlug(ctx, slug)
}
func (m *mockService) PublishContentItem(ctx context.Context, contentItemID string) error {
	if m.publishContentItem == nil {
		panic("mockService.PublishContentItem not set")
	}
	return m.publishContentItem(ctx, contentItemID)
}
func (m *mockService) UpdateContentItem(ctx context.Context, req contentdomain.UpdateContentItemRequest) error {
	if m.updateContentItem == nil {
		panic("mockService.UpdateContentItem not set")
	}
	return m.updateContentItem(ctx, req)
}
func (m *mockService) GetAllContentItems(ctx context.Context, getType contentdomain.ContentItemStatus, params pagination.Params) (contentItemList []*contentdomain.ContentItem, err error) {
	if m.getAllContentItems == nil {
		panic("mockService.GetAllContentItems not set")
	}
	return m.getAllContentItems(ctx, getType, params)
}
