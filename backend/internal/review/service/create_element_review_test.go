package reviewservice

import (
	"context"
	"errors"
	reviewdomain "learnflow_backend/internal/review/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCourseReviewAccess(t *testing.T) {
	Convey("Create CourseReview access checks", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessCourse: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(cRepo, nil, accessChecker)

		Convey("No permission", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("access error", func() {
			accessChecker.hasAccessCourse = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestCreateCourseReviewLookup(t *testing.T) {
	Convey("Create CourseReview — existing-review lookup", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessCourse: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(cRepo, nil, accessChecker)

		Convey("GetCourseReviewByUserAndCourseID error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetCourseReviewByUserAndCourseID collision", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateCourseReviewPersistence(t *testing.T) {
	Convey("Create CourseReview persistence", t, func() {
		cRepo := &mockReviewRepo{
			getCourseReviewByUserAndCourseID: func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		accessChecker := &mockAccessChecker{
			hasAccessCourse: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(cRepo, nil, accessChecker)

		Convey("CreateCourseReview error", func() {
			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReview(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateContentReviewAccess(t *testing.T) {
	Convey("Create ContentReview access checks", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessContent: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(nil, cRepo, accessChecker)

		Convey("No permission", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrNoPermission), ShouldBeTrue)
		})

		Convey("access error", func() {
			accessChecker.hasAccessContent = func(_ context.Context, _ string, _ string) (bool, error) {
				return false, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewLookup(t *testing.T) {
	Convey("Create ContentReview — existing-review lookup", t, func() {
		cRepo := &mockReviewRepo{}
		accessChecker := &mockAccessChecker{
			hasAccessContent: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(nil, cRepo, accessChecker)

		Convey("GetContentReviewByUserAndContentID error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetContentReviewByUserAndContentID collision", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewPersistence(t *testing.T) {
	Convey("Create ContentReview persistence", t, func() {
		cRepo := &mockReviewRepo{
			getContentReviewByUserAndContentID: func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		accessChecker := &mockAccessChecker{
			hasAccessContent: func(_ context.Context, _ string, _ string) (bool, error) {
				return true, nil
			},
		}
		srv := newTestService(nil, cRepo, accessChecker)

		Convey("CreateContentReview error", func() {
			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReview(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateCourseReviewAdminLookup(t *testing.T) {
	Convey("Create CourseReview as admin — existing-review lookup", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(cRepo, nil, nil)

		Convey("GetCourseReviewByUserAndCourseID error", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetCourseReviewByUserAndCourseID collision", func() {
			cRepo.getCourseReviewByUserAndCourseID = func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateCourseReviewAdminPersistence(t *testing.T) {
	Convey("Create CourseReview as admin — persistence", t, func() {
		cRepo := &mockReviewRepo{
			getCourseReviewByUserAndCourseID: func(_ context.Context, _ string, _ string) (*reviewdomain.CourseReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		srv := newTestService(cRepo, nil, nil)

		Convey("CreateCourseReview error", func() {
			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.createCourseReview = func(_ context.Context, _ *reviewdomain.CourseReview) (*reviewdomain.CourseReview, error) {
				return &reviewdomain.CourseReview{}, nil
			}

			err := srv.CreateCourseReviewAdmin(context.Background(), reviewdomain.CreateCourseReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateContentReviewAdminLookup(t *testing.T) {
	Convey("Create ContentReview as admin — existing-review lookup", t, func() {
		cRepo := &mockReviewRepo{}
		srv := newTestService(nil, cRepo, nil)

		Convey("GetContentReviewByUserAndContentID error", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetContentReviewByUserAndContentID collision", func() {
			cRepo.getContentReviewByUserAndContentID = func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateContentReviewAdminPersistence(t *testing.T) {
	Convey("Create ContentReview as admin — persistence", t, func() {
		cRepo := &mockReviewRepo{
			getContentReviewByUserAndContentID: func(_ context.Context, _ string, _ string) (*reviewdomain.ContentReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		srv := newTestService(nil, cRepo, nil)

		Convey("CreateContentReview error", func() {
			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			cRepo.createContentReview = func(_ context.Context, _ *reviewdomain.ContentReview) (*reviewdomain.ContentReview, error) {
				return &reviewdomain.ContentReview{}, nil
			}

			err := srv.CreateContentReviewAdmin(context.Background(), reviewdomain.CreateContentReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateArticleReviewLookup(t *testing.T) {
	Convey("Create ArticleReview — existing-review lookup", t, func() {
		aRepo := &mockReviewRepo{}
		srv := newTestServiceWithArticleRepo(aRepo)

		Convey("GetArticleReviewByUserAndArticleID error", func() {
			aRepo.getArticleReviewByUserAndArticleID = func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateArticleReview(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetArticleReviewByUserAndArticleID collision", func() {
			aRepo.getArticleReviewByUserAndArticleID = func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{}, nil
			}

			err := srv.CreateArticleReview(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateArticleReviewPersistence(t *testing.T) {
	Convey("Create ArticleReview — persistence", t, func() {
		aRepo := &mockReviewRepo{
			getArticleReviewByUserAndArticleID: func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		srv := newTestServiceWithArticleRepo(aRepo)

		Convey("CreateArticleReview error", func() {
			aRepo.createArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateArticleReview(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			aRepo.createArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{}, nil
			}

			err := srv.CreateArticleReview(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateArticleReviewAdminLookup(t *testing.T) {
	Convey("Create ArticleReview as admin — existing-review lookup", t, func() {
		aRepo := &mockReviewRepo{}
		srv := newTestServiceWithArticleRepo(aRepo)

		Convey("GetArticleReviewByUserAndArticleID error", func() {
			aRepo.getArticleReviewByUserAndArticleID = func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateArticleReviewAdmin(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("GetArticleReviewByUserAndArticleID collision", func() {
			aRepo.getArticleReviewByUserAndArticleID = func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{}, nil
			}

			err := srv.CreateArticleReviewAdmin(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, reviewdomain.ErrAlreadyReviewed), ShouldBeTrue)
		})
	})
}

func TestCreateArticleReviewAdminPersistence(t *testing.T) {
	Convey("Create ArticleReview as admin — persistence", t, func() {
		aRepo := &mockReviewRepo{
			getArticleReviewByUserAndArticleID: func(_ context.Context, _ string, _ string) (*reviewdomain.ArticleReview, error) {
				return nil, reviewdomain.ErrReviewNotFound
			},
		}
		srv := newTestServiceWithArticleRepo(aRepo)

		Convey("CreateArticleReview error", func() {
			aRepo.createArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
				return nil, testutil.ErrDBUnexpected
			}

			err := srv.CreateArticleReviewAdmin(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("success", func() {
			aRepo.createArticleReview = func(_ context.Context, _ *reviewdomain.ArticleReview) (*reviewdomain.ArticleReview, error) {
				return &reviewdomain.ArticleReview{}, nil
			}

			err := srv.CreateArticleReviewAdmin(context.Background(), reviewdomain.CreateArticleReviewRequest{})
			So(err, ShouldBeNil)
		})
	})
}
