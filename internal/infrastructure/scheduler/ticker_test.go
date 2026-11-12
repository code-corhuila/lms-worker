package scheduler_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/code-corhuila/lms-worker/internal/infrastructure/scheduler"
)

const testTimeout = 5 * time.Second

// fakeJob — Fake test double (11-quality/tdd-guide.md). It announces every run
// on the runs channel without ever blocking the ticker.
type fakeJob struct {
	err   error
	count atomic.Int32
	runs  chan struct{}
}

func newFakeJob(err error) *fakeJob {
	return &fakeJob{err: err, runs: make(chan struct{}, 100)}
}

func (j *fakeJob) Execute(_ context.Context) error {
	j.count.Add(1)
	select {
	case j.runs <- struct{}{}:
	default:
	}
	return j.err
}

func (j *fakeJob) waitForRuns(t *testing.T, n int) {
	t.Helper()
	for range n {
		select {
		case <-j.runs:
		case <-time.After(testTimeout):
			t.Fatalf("job ran fewer than %d times within %s", n, testTimeout)
		}
	}
}

// startTicker runs the ticker in the background and returns a stop function
// that cancels it and waits until Run has returned.
func startTicker(t *testing.T, interval time.Duration, jobs map[string]scheduler.Job, log *zap.SugaredLogger) (stop func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		scheduler.NewTicker(interval, jobs, log).Run(ctx)
	}()
	return func() {
		t.Helper()
		cancel()
		select {
		case <-done:
		case <-time.After(testTimeout):
			t.Fatal("Run did not return after the context was cancelled")
		}
	}
}

func TestTicker_RunsEveryJobImmediately(t *testing.T) {
	first, second := newFakeJob(nil), newFakeJob(nil)
	jobs := map[string]scheduler.Job{"first": first, "second": second}

	stop := startTicker(t, time.Hour, jobs, zap.NewNop().Sugar())
	first.waitForRuns(t, 1)
	second.waitForRuns(t, 1)
	stop()

	if got := first.count.Load(); got != 1 {
		t.Errorf("first ran %d times before the interval elapsed, want 1", got)
	}
}

func TestTicker_RepeatsOnEveryInterval(t *testing.T) {
	job := newFakeJob(nil)

	stop := startTicker(t, time.Millisecond, map[string]scheduler.Job{"job": job}, zap.NewNop().Sugar())
	job.waitForRuns(t, 3)
	stop()
}

func TestTicker_StopsWhenContextIsCancelled(t *testing.T) {
	job := newFakeJob(nil)

	stop := startTicker(t, time.Millisecond, map[string]scheduler.Job{"job": job}, zap.NewNop().Sugar())
	job.waitForRuns(t, 1)
	stop()

	runsAtStop := job.count.Load()
	time.Sleep(20 * time.Millisecond)
	if got := job.count.Load(); got != runsAtStop {
		t.Errorf("job ran %d more times after Run returned", got-runsAtStop)
	}
}

func TestTicker_FailingJobDoesNotAffectTheOthers(t *testing.T) {
	failing, healthy := newFakeJob(errors.New("upstream down")), newFakeJob(nil)
	jobs := map[string]scheduler.Job{"failing": failing, "healthy": healthy}
	core, logs := observer.New(zapcore.ErrorLevel)

	stop := startTicker(t, time.Hour, jobs, zap.New(core).Sugar())
	failing.waitForRuns(t, 1)
	healthy.waitForRuns(t, 1)
	stop()

	entries := logs.FilterMessage("job failed").All()
	if len(entries) != 1 {
		t.Fatalf("error logs = %d, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["job"]; got != "failing" {
		t.Errorf("logged job = %v, want failing", got)
	}
}
