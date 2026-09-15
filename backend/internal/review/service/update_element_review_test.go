package reviewservice

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateCourseReview(t *testing.T) {
	Convey("Update CourseReview", t, func() {
		cRepo := &mockReviewRepo{
			getCourseReviewByID: func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{ID: "review-1", CourseID: "course-1", UserID: "user-1"}, nil
			},
		}
		accessChecker := &mockAccessChecker{
			hasAccessCourse: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(cRepo, nil, accessChecker)
		req := reviewdomain.UpdateCourseReviewRequest{ReviewID: "review-1", UserID: "user-1"}

		Convey("fetch review error", func() {
			cRepo.getCourseReviewByID = func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("not the review owner", func() {
			cRepo.getCourseReviewByID = func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{ID: "review-1", CourseID: "course-1", UserID: "someone-else"}, nil
			}

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("access check error", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("no permission", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			cRepo.updateCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			var applied *reviewdomain.CourseReview
			cRepo.updateCourseReview = func(_ context.Context, r *reviewdomain.CourseReview) error {
				applied = r
				return nil
			}
			rating := 5
			req.Rating = &rating

			err := srv.UpdateCourseReview(context.Background(), req)
			So(err, ShouldBeNil)
			So(applied.Rating, ShouldEqual, 5)
		})
	})
}

func TestUpdateContentReview(t *testing.T) {
	Convey("Update ContentReview", t, func() {
		cRepo := &mockReviewRepo{
			getContentReviewByID: func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{ID: "review-1", ContentID: "content-1", UserID: "user-1"}, nil
			},
		}
		accessChecker := &mockAccessChecker{
			hasAccessContent: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(nil, cRepo, accessChecker)
		req := reviewdomain.UpdateContentReviewRequest{ReviewID: "review-1", UserID: "user-1"}

		Convey("fetch review error", func() {
			cRepo.getContentReviewByID = func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("not the review owner", func() {
			cRepo.getContentReviewByID = func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{ID: "review-1", ContentID: "content-1", UserID: "someone-else"}, nil
			}

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("access check error", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("no permission", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			cRepo.updateContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			var applied *reviewdomain.ContentReview
			cRepo.updateContentReview = func(_ context.Context, r *reviewdomain.ContentReview) error {
				applied = r
				return nil
			}
			rating := 4
			req.Rating = &rating

			err := srv.UpdateContentReview(context.Background(), req)
			So(err, ShouldBeNil)
			So(applied.Rating, ShouldEqual, 4)
		})
	})
}

func TestUpdateCourseReviewAdmin(t *testing.T) {
	Convey("Update CourseReview as admin", t, func() {
		cRepo := &mockReviewRepo{
			getCourseReviewByID: func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{ID: "review-1", CourseID: "course-1", UserID: "user-1"}, nil
			},
		}
		srv := newTestService(cRepo, nil, nil)
		req := reviewdomain.UpdateCourseReviewRequest{ReviewID: "review-1"}

		Convey("fetch review error", func() {
			cRepo.getCourseReviewByID = func(_ context.Context, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateCourseReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			cRepo.updateCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateCourseReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.updateCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) error {
				return nil
			}

			err := srv.UpdateCourseReviewAdmin(context.Background(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateContentReviewAdmin(t *testing.T) {
	Convey("Update ContentReview as admin", t, func() {
		cRepo := &mockReviewRepo{
			getContentReviewByID: func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{ID: "review-1", ContentID: "content-1", UserID: "user-1"}, nil
			},
		}
		srv := newTestService(nil, cRepo, nil)
		req := reviewdomain.UpdateContentReviewRequest{ReviewID: "review-1"}

		Convey("fetch review error", func() {
			cRepo.getContentReviewByID = func(_ context.Context, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateContentReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			cRepo.updateContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateContentReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.updateContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) error {
				return nil
			}

			err := srv.UpdateContentReviewAdmin(context.Background(), req)
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateArticleReview(t *testing.T) {
	Convey("Update ArticleReview", t, func() {
		aRepo := &mockReviewRepo{
			getArticleReviewByID: func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{ID: "review-1", ArticleID: "article-1", UserID: "user-1"}, nil
			},
		}
		srv := newTestServiceWithArticleRepo(aRepo)
		req := reviewdomain.UpdateArticleReviewRequest{ReviewID: "review-1", UserID: "user-1"}

		Convey("fetch review error", func() {
			aRepo.getArticleReviewByID = func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateArticleReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("not the review owner", func() {
			aRepo.getArticleReviewByID = func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{ID: "review-1", ArticleID: "article-1", UserID: "someone-else"}, nil
			}

			err := srv.UpdateArticleReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrReviewNotFound), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			aRepo.updateArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateArticleReview(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			var applied *reviewdomain.ArticleReview
			aRepo.updateArticleReview = func(_ context.Context, r *reviewdomain.ArticleReview) error {
				applied = r
				return nil
			}
			rating := 5
			req.Rating = &rating

			err := srv.UpdateArticleReview(context.Background(), req)
			So(err, ShouldBeNil)
			So(applied.Rating, ShouldEqual, 5)
		})
	})
}

func TestUpdateArticleReviewAdmin(t *testing.T) {
	Convey("Update ArticleReview as admin", t, func() {
		aRepo := &mockReviewRepo{
			getArticleReviewByID: func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{ID: "review-1", ArticleID: "article-1", UserID: "user-1"}, nil
			},
		}
		srv := newTestServiceWithArticleRepo(aRepo)
		req := reviewdomain.UpdateArticleReviewRequest{ReviewID: "review-1"}

		Convey("fetch review error", func() {
			aRepo.getArticleReviewByID = func(_ context.Context, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.UpdateArticleReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("repository update error", func() {
			aRepo.updateArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) error {
				return testutil.ErrDBUnexpected
			}

			err := srv.UpdateArticleReviewAdmin(context.Background(), req)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			aRepo.updateArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) error {
				return nil
			}

			err := srv.UpdateArticleReviewAdmin(context.Background(), req)
			So(err, ShouldBeNil)
		})
	})
}
