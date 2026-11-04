package usecase_test

import (
	"context"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

var (
	_ notice.CirculationClient = (*fakeCirculationClient)(nil)
	_ notice.CatalogClient     = (*fakeCatalogClient)(nil)
)

// fakeCirculationClient — Fake test double (11-quality/tdd-guide.md).
type fakeCirculationClient struct {
	loans      []notice.OverdueLoan
	findErr    error
	markErrs   map[string]error
	accrueErr  error
	marked     []string
	accrueRuns int
}

func (f *fakeCirculationClient) FindLoansNeedingNotice(_ context.Context) ([]notice.OverdueLoan, error) {
	return f.loans, f.findErr
}

func (f *fakeCirculationClient) MarkNotified(_ context.Context, loanID string) error {
	f.marked = append(f.marked, loanID)
	return f.markErrs[loanID]
}

func (f *fakeCirculationClient) AccrueFines(_ context.Context) error {
	f.accrueRuns++
	return f.accrueErr
}

// fakeCatalogClient — Fake test double (11-quality/tdd-guide.md).
type fakeCatalogClient struct {
	reindexErr  error
	reindexRuns int
}

func (f *fakeCatalogClient) Reindex(_ context.Context) error {
	f.reindexRuns++
	return f.reindexErr
}
