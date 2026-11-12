package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

type CirculationClient struct {
	baseURL   string
	jwtSecret string
	http      *http.Client
}

func NewCirculationClient(baseURL, jwtSecret string) *CirculationClient {
	return &CirculationClient{baseURL: baseURL, jwtSecret: jwtSecret, http: &http.Client{Timeout: 5 * time.Second}}
}

// overdueLoanResponse mirrors lms-circulation-api's existing loanResponse
// shape (internal/adapter/in/httpapi/handler/loan_handler.go) — `id`, not
// `loanId`; only the fields this job actually needs.
//
// NOTE: `notifiedAt`-based filtering (so this endpoint returns only loans
// that still need a notice) is new surface area lms-circulation-api doesn't
// expose yet — this client calls the existing GET /api/v1/loans/overdue and
// relies on that filtering being added there; flagged in this repo's README.
type overdueLoanResponse struct {
	ID        string `json:"id"`
	StudentID string `json:"studentId"`
}

type overdueLoansEnvelope struct {
	Data []overdueLoanResponse `json:"data"`
}

func (c *CirculationClient) FindLoansNeedingNotice(ctx context.Context) ([]notice.OverdueLoan, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/loans/overdue")
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling circulation-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("circulation-service returned %d", resp.StatusCode)
	}

	var body overdueLoansEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding circulation-service response: %w", err)
	}

	loans := make([]notice.OverdueLoan, 0, len(body.Data))
	for _, l := range body.Data {
		loans = append(loans, notice.OverdueLoan{LoanID: l.ID, StudentID: l.StudentID})
	}
	return loans, nil
}

// MarkNotified: new endpoint, not yet built on lms-circulation-api — see
// this repo's README.
func (c *CirculationClient) MarkNotified(ctx context.Context, loanID string) error {
	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/loans/%s/notify", loanID))
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling circulation-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("circulation-service returned %d", resp.StatusCode)
	}
	return nil
}

// AccrueFines: new endpoint, not yet built on lms-circulation-api — see
// this repo's README.
func (c *CirculationClient) AccrueFines(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/api/v1/loans/accrue-fines")
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling circulation-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("circulation-service returned %d", resp.StatusCode)
	}
	return nil
}

func (c *CirculationClient) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	token, err := mintServiceToken(c.jwtSecret, "lms-worker")
	if err != nil {
		return nil, fmt.Errorf("minting internal service token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}
