package usecase_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/code-corhuila/lms-worker/internal/application/usecase"
	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

func TestDueDateNotices_Execute(t *testing.T) {
	errUpstream := errors.New("circulation unavailable")
	loans := []notice.OverdueLoan{
		{LoanID: "loan-1", StudentID: "student-1"},
		{LoanID: "loan-2", StudentID: "student-2"},
	}

	tests := []struct {
		name        string
		loans       []notice.OverdueLoan
		findErr     error
		markErrs    map[string]error
		wantErr     error
		wantMarked  []string
		wantNotices int
		wantErrLogs int
	}{
		{
			name: "no overdue loans",
		},
		{
			name:        "notifies and marks every overdue loan",
			loans:       loans,
			wantMarked:  []string{"loan-1", "loan-2"},
			wantNotices: 2,
		},
		{
			name:    "find failure is returned and nothing is marked",
			findErr: errUpstream,
			wantErr: errUpstream,
		},
		{
			name:        "mark failure is logged and does not stop the remaining loans",
			loans:       loans,
			markErrs:    map[string]error{"loan-1": errUpstream},
			wantMarked:  []string{"loan-1", "loan-2"},
			wantNotices: 2,
			wantErrLogs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circulation := &fakeCirculationClient{loans: tt.loans, findErr: tt.findErr, markErrs: tt.markErrs}
			core, logs := observer.New(zapcore.InfoLevel)
			uc := usecase.NewDueDateNotices(circulation, zap.New(core).Sugar())

			err := uc.Execute(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantErr)
			}
			if !slices.Equal(circulation.marked, tt.wantMarked) {
				t.Errorf("marked = %v, want %v", circulation.marked, tt.wantMarked)
			}
			if got := logs.FilterMessage("due-date notice").Len(); got != tt.wantNotices {
				t.Errorf("notices logged = %d, want %d", got, tt.wantNotices)
			}
			if got := logs.FilterLevelExact(zapcore.ErrorLevel).Len(); got != tt.wantErrLogs {
				t.Errorf("error logs = %d, want %d", got, tt.wantErrLogs)
			}
		})
	}
}
