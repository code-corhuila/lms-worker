// Package notice holds the driven ports for lms-worker's three jobs
// (ADR-009-worker-scheduling-model.md: due-date notices, fines accrual,
// catalog reindexing). This service has no business logic of its own — it
// calls existing idempotent operations on lms-circulation-api/
// lms-catalog-api and does not need to reimplement their guarantees
// (ADR-009's own justification). There is deliberately no domain aggregate
// here, only the ports a scheduled job needs.
package notice

import "context"

// OverdueLoan is the shape this service needs from lms-circulation-api's
// own query — it never reads that service's database directly
// (library-docs/09-microservices/service-boundary-rules.md).
type OverdueLoan struct {
	LoanID    string
	StudentID string
}

// CirculationClient is the driven port for everything this service needs
// from lms-circulation-api.
type CirculationClient interface {
	// FindLoansNeedingNotice returns overdue loans that have not yet been
	// notified — the dedupe marker (`notifiedAt`) lives on the loan record
	// itself in lms-circulation-api, not in this service's own storage, so
	// this service stays genuinely stateless (ADR-006, ADR-009).
	FindLoansNeedingNotice(ctx context.Context) ([]OverdueLoan, error)
	// MarkNotified sets that marker once a notice has been sent for a loan.
	MarkNotified(ctx context.Context, loanID string) error
	// AccrueFines triggers lms-circulation-api's own idempotent fine-accrual
	// operation — this service only decides *when*, never *how*.
	AccrueFines(ctx context.Context) error
}

// CatalogClient is the driven port for lms-catalog-api's reindex trigger.
type CatalogClient interface {
	Reindex(ctx context.Context) error
}
