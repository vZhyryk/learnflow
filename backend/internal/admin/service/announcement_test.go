package adminservice

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateAnnouncement(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)

		Convey("When the repository returns an error", func() {
			repo.createAnnouncement = func(_ context.Context, _ *admindomain.Announcement) (*admindomain.Announcement, error) {
				return nil, testutil.ErrDBUnexpected
			}

			id, err := srv.CreateAnnouncement(context.Background(), admindomain.CreateAnnouncementRequest{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(id, ShouldBeEmpty)
		})

		Convey("When it succeeds", func() {
			var got *admindomain.Announcement
			repo.createAnnouncement = func(_ context.Context, announcement *admindomain.Announcement) (*admindomain.Announcement, error) {
				got = announcement
				return &admindomain.Announcement{ID: "announcement-123"}, nil
			}

			id, err := srv.CreateAnnouncement(context.Background(), admindomain.CreateAnnouncementRequest{Title: "T", CreatedByUserID: "user-1"})
			So(err, ShouldBeNil)
			So(id, ShouldEqual, "announcement-123")
			So(got.CreatedByUserID, ShouldEqual, "user-1")
		})
	})
}

func getValidAnnouncement(_ context.Context, _ string) (*admindomain.Announcement, error) {
	entityID := "course-123"
	entityType := admindomain.CourseEntityType
	return &admindomain.Announcement{
		ID:         "announcement-123",
		Title:      "old title",
		EntityID:   &entityID,
		EntityType: &entityType,
	}, nil
}

func TestUpdateAnnouncement(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{
			getAnnouncementByID: getValidAnnouncement,
		}
		srv := newTestService(repo)

		Convey("When fetching the existing announcement fails", func() {
			repo.getAnnouncementByID = func(_ context.Context, _ string) (*admindomain.Announcement, error) {
				return nil, admindomain.ErrAnnouncementNotFound
			}

			err := srv.UpdateAnnouncement(context.Background(), admindomain.UpdateAnnouncementRequest{ID: "announcement-123"})
			So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
		})

		Convey("When the resulting announcement has entityID set but no entityType", func() {
			repo.getAnnouncementByID = func(_ context.Context, _ string) (*admindomain.Announcement, error) {
				entityID := "course-123"
				return &admindomain.Announcement{ID: "announcement-123", EntityID: &entityID, EntityType: nil}, nil
			}

			err := srv.UpdateAnnouncement(context.Background(), admindomain.UpdateAnnouncementRequest{ID: "announcement-123"})
			So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
		})

		Convey("When the announcement is already approved", func() {
			approvedAt := time.Now()
			repo.getAnnouncementByID = func(_ context.Context, _ string) (*admindomain.Announcement, error) {
				return &admindomain.Announcement{ID: "announcement-123", ApprovedAt: &approvedAt}, nil
			}

			err := srv.UpdateAnnouncement(context.Background(), admindomain.UpdateAnnouncementRequest{ID: "announcement-123"})
			So(errors.Is(err, admindomain.ErrAnnouncementApproved), ShouldBeTrue)
		})

		Convey("When the repository update fails", func() {
			repo.updateAnnouncement = func(_ context.Context, _ *admindomain.Announcement) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateAnnouncement(context.Background(), admindomain.UpdateAnnouncementRequest{ID: "announcement-123"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			var applied *admindomain.Announcement
			repo.updateAnnouncement = func(_ context.Context, a *admindomain.Announcement) error {
				applied = a
				return nil
			}
			newTitle := "new title"

			err := srv.UpdateAnnouncement(context.Background(), admindomain.UpdateAnnouncementRequest{
				ID: "announcement-123", Title: &newTitle,
			})
			So(err, ShouldBeNil)
			So(applied.Title, ShouldEqual, "new title")
		})
	})
}

func TestApproveAnnouncement(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)

		Convey("When the repository returns an error", func() {
			repo.approveAnnouncement = testutil.AlwaysFailsDB2
			err := srv.ApproveAnnouncement(context.Background(), "announcement-123", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			repo.approveAnnouncement = testutil.AlwaysNil2
			err := srv.ApproveAnnouncement(context.Background(), "announcement-123", "user-1")
			So(err, ShouldBeNil)
		})
	})
}

func TestGetAnnouncements(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)
		params := pagination.NewParams(1, 20)

		Convey("When the repository returns an error", func() {
			repo.getAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return nil, testutil.ErrDBUnexpected
			}
			_, err := srv.GetAnnouncements(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			want := []*admindomain.Announcement{{ID: "announcement-1"}}
			repo.getAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return want, nil
			}
			got, err := srv.GetAnnouncements(context.Background(), params)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}

func TestGetUnApprovedAnnouncements(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)
		params := pagination.NewParams(1, 20)

		Convey("When the repository returns an error", func() {
			repo.getUnApprovedAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return nil, testutil.ErrDBUnexpected
			}
			_, err := srv.GetUnApprovedAnnouncements(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			want := []*admindomain.Announcement{{ID: "announcement-1"}}
			repo.getUnApprovedAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return want, nil
			}
			got, err := srv.GetUnApprovedAnnouncements(context.Background(), params)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}

func TestGetApprovedAnnouncements(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)
		params := pagination.NewParams(1, 20)

		Convey("When the repository returns an error", func() {
			repo.getApprovedAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return nil, testutil.ErrDBUnexpected
			}
			_, err := srv.GetApprovedAnnouncements(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			want := []*admindomain.Announcement{{ID: "announcement-1"}}
			repo.getApprovedAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return want, nil
			}
			got, err := srv.GetApprovedAnnouncements(context.Background(), params)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}

func TestGetPublicAnnouncements(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)
		params := pagination.NewParams(1, 20)

		Convey("When the repository returns an error", func() {
			repo.getPublicAnnouncements = func(_ context.Context, _ pagination.Params, _ string) ([]*admindomain.AnnouncementPublic, error) {
				return nil, testutil.ErrDBUnexpected
			}
			_, err := srv.GetPublicAnnouncements(context.Background(), params, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.GetPublicAnnouncements")
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds, the user ID and params reach the repository", func() {
			var gotUserID string
			var gotParams pagination.Params
			want := []*admindomain.AnnouncementPublic{{ID: "announcement-1"}}
			repo.getPublicAnnouncements = func(_ context.Context, p pagination.Params, userID string) ([]*admindomain.AnnouncementPublic, error) {
				gotParams, gotUserID = p, userID
				return want, nil
			}
			got, err := srv.GetPublicAnnouncements(context.Background(), params, "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(gotUserID, ShouldEqual, "user-1")
			So(gotParams, ShouldResemble, params)
		})
	})
}

func TestGetExpiredAnnouncements(t *testing.T) {
	Convey("Given an admin service", t, func() {
		repo := &mockAnnouncementRepo{}
		srv := newTestService(repo)
		params := pagination.NewParams(1, 20)

		Convey("When the repository returns an error", func() {
			repo.getExpiredAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return nil, testutil.ErrDBUnexpected
			}
			_, err := srv.GetExpiredAnnouncements(context.Background(), params)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When it succeeds", func() {
			want := []*admindomain.Announcement{{ID: "announcement-1"}}
			repo.getExpiredAnnouncements = func(_ context.Context, _ pagination.Params) ([]*admindomain.Announcement, error) {
				return want, nil
			}
			got, err := srv.GetExpiredAnnouncements(context.Background(), params)
			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
		})
	})
}
