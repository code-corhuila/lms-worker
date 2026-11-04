package usecase

import (
	"context"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

// FinesAccrual implements ADR-009's second responsibility: trigger
// lms-circulation-api's own idempotent fine-accrual operation on a
// schedule. This job decides *when*; lms-circulation-api decides *how* and
// guarantees the operation is safe to call repeatedly.
type FinesAccrual struct {
	Circulation notice.CirculationClient
}

func NewFinesAccrual(circulation notice.CirculationClient) *FinesAccrual {
	return &FinesAccrual{Circulation: circulation}
}

func (uc *FinesAccrual) Execute(ctx context.Context) error {
	return uc.Circulation.AccrueFines(ctx)
}
