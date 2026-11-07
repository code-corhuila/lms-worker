package usecase

import (
	"context"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

// CatalogReindex implements ADR-009's third responsibility: trigger
// lms-catalog-api's search-index reindex on a schedule.
type CatalogReindex struct {
	Catalog notice.CatalogClient
}

func NewCatalogReindex(catalog notice.CatalogClient) *CatalogReindex {
	return &CatalogReindex{Catalog: catalog}
}

func (uc *CatalogReindex) Execute(ctx context.Context) error {
	return uc.Catalog.Reindex(ctx)
}
