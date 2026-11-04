package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/code-corhuila/lms-worker/internal/application/usecase"
)

func TestCatalogReindex_Execute(t *testing.T) {
	errUpstream := errors.New("catalog unavailable")

	tests := []struct {
		name       string
		reindexErr error
	}{
		{name: "triggers the reindex once"},
		{name: "reindex failure is returned", reindexErr: errUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := &fakeCatalogClient{reindexErr: tt.reindexErr}
			uc := usecase.NewCatalogReindex(catalog)

			err := uc.Execute(context.Background())

			if !errors.Is(err, tt.reindexErr) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.reindexErr)
			}
			if catalog.reindexRuns != 1 {
				t.Errorf("reindex runs = %d, want 1", catalog.reindexRuns)
			}
		})
	}
}
