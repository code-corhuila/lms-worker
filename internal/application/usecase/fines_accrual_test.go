package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/code-corhuila/lms-worker/internal/application/usecase"
)

func TestFinesAccrual_Execute(t *testing.T) {
	errUpstream := errors.New("circulation unavailable")

	tests := []struct {
		name      string
		accrueErr error
	}{
		{name: "triggers the accrual once"},
		{name: "accrual failure is returned", accrueErr: errUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circulation := &fakeCirculationClient{accrueErr: tt.accrueErr}
			uc := usecase.NewFinesAccrual(circulation)

			err := uc.Execute(context.Background())

			if !errors.Is(err, tt.accrueErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.accrueErr)
			}
			if circulation.accrueRuns != 1 {
				t.Errorf("accrual runs = %d, want 1", circulation.accrueRuns)
			}
		})
	}
}
