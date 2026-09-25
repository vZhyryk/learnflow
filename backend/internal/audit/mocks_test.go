package audit

import (
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Audit {
	return &Audit{BaseRepository: repository.BaseRepository{DB: runner}}
}
