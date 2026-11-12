// Package scheduler drives the three jobs on ADR-009's polling cadence —
// not event-driven in the message-broker sense (AT-002); a plain
// time.Ticker, same as the course's "Minimal Infrastructure Footprint"
// principle (P5) elsewhere.
package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Job is any of the three use cases — kept to this single-method shape so
// the ticker can drive them uniformly and in isolation from one another.
type Job interface {
	Execute(ctx context.Context) error
}

type Ticker struct {
	interval time.Duration
	jobs     map[string]Job
	log      *zap.SugaredLogger
}

func NewTicker(interval time.Duration, jobs map[string]Job, log *zap.SugaredLogger) *Ticker {
	return &Ticker{interval: interval, jobs: jobs, log: log}
}

// Run blocks until ctx is cancelled, driving one tick immediately and then
// every t.interval.
func (t *Ticker) Run(ctx context.Context) {
	t.tick(ctx)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

// tick runs every job independently — one job failing (ADR-009's "per-tick
// isolation") never blocks or skips the others.
func (t *Ticker) tick(ctx context.Context) {
	for name, job := range t.jobs {
		if err := job.Execute(ctx); err != nil {
			t.log.Errorw("job failed", "job", name, "error", err)
		}
	}
}
