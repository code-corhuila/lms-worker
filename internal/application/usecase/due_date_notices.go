// Package usecase orchestrates each of lms-worker's three jobs — no
// business logic here (that lives in the services this worker calls), only
// coordination (library-docs/05-architecture/hexagonal-architecture.md).
package usecase

import (
	"context"

	"go.uber.org/zap"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

// DueDateNotices implements ADR-009's first responsibility: find overdue
// loans that have not yet been notified, send the notice, mark them
// notified. "Send" here means log it — there is no email/SMS integration in
// this system yet; the dedupe/idempotency contract is what this job
// actually exists to get right, independent of the notification channel.
type DueDateNotices struct {
	Circulation notice.CirculationClient
	Log         *zap.SugaredLogger
}

func NewDueDateNotices(circulation notice.CirculationClient, log *zap.SugaredLogger) *DueDateNotices {
	return &DueDateNotices{Circulation: circulation, Log: log}
}

func (uc *DueDateNotices) Execute(ctx context.Context) error {
	loans, err := uc.Circulation.FindLoansNeedingNotice(ctx)
	if err != nil {
		return err
	}

	for _, loan := range loans {
		uc.Log.Infow("due-date notice", "loanId", loan.LoanID, "studentId", loan.StudentID)
		if err := uc.Circulation.MarkNotified(ctx, loan.LoanID); err != nil {
			uc.Log.Errorw("marking loan notified", "loanId", loan.LoanID, "error", err)
		}
	}
	return nil
}
