package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type CatalogClient struct {
	baseURL   string
	jwtSecret string
	http      *http.Client
}

func NewCatalogClient(baseURL, jwtSecret string) *CatalogClient {
	return &CatalogClient{baseURL: baseURL, jwtSecret: jwtSecret, http: &http.Client{Timeout: 5 * time.Second}}
}

// Reindex: new endpoint, not yet built on lms-catalog-api — see this repo's
// README.
func (c *CatalogClient) Reindex(ctx context.Context) error {
	token, err := mintServiceToken(c.jwtSecret, "lms-worker")
	if err != nil {
		return fmt.Errorf("minting internal service token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/books/reindex", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling catalog-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("catalog-service returned %d", resp.StatusCode)
	}
	return nil
}
