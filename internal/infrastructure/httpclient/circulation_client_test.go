package httpclient

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/code-corhuila/lms-worker/internal/domain/notice"
)

func TestCirculationClient_FindLoansNeedingNotice(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantLoans []notice.OverdueLoan
		wantErr   string
	}{
		{
			name:   "maps the response envelope to overdue loans",
			status: http.StatusOK,
			body:   `{"data":[{"id":"loan-1","studentId":"student-1"},{"id":"loan-2","studentId":"student-2"}]}`,
			wantLoans: []notice.OverdueLoan{
				{LoanID: "loan-1", StudentID: "student-1"},
				{LoanID: "loan-2", StudentID: "student-2"},
			},
		},
		{name: "no overdue loans", status: http.StatusOK, body: `{"data":[]}`},
		{name: "non-200 status is an error", status: http.StatusInternalServerError, wantErr: "returned 500"},
		{name: "malformed body is an error", status: http.StatusOK, body: `{`, wantErr: "decoding"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, requests := newFakeServer(t, tt.status, tt.body)
			client := NewCirculationClient(srv.URL, testSecret)

			loans, err := client.FindLoansNeedingNotice(context.Background())

			assertErrContains(t, err, tt.wantErr)
			assertRequest(t, requests, http.MethodGet, "/api/v1/loans/overdue")
			if !slices.Equal(loans, tt.wantLoans) {
				t.Errorf("loans = %v, want %v", loans, tt.wantLoans)
			}
		})
	}
}

func TestCirculationClient_Commands(t *testing.T) {
	commands := []struct {
		name string
		path string
		call func(*CirculationClient) error
	}{
		{
			name: "MarkNotified",
			path: "/api/v1/loans/loan-1/notify",
			call: func(c *CirculationClient) error { return c.MarkNotified(context.Background(), "loan-1") },
		},
		{
			name: "AccrueFines",
			path: "/api/v1/loans/accrue-fines",
			call: func(c *CirculationClient) error { return c.AccrueFines(context.Background()) },
		},
	}
	statuses := []struct {
		name    string
		status  int
		wantErr string
	}{
		{name: "ok", status: http.StatusOK},
		{name: "not found", status: http.StatusNotFound, wantErr: "returned 404"},
		{name: "server error", status: http.StatusInternalServerError, wantErr: "returned 500"},
	}

	for _, cmd := range commands {
		for _, st := range statuses {
			t.Run(cmd.name+"/"+st.name, func(t *testing.T) {
				srv, requests := newFakeServer(t, st.status, "")
				client := NewCirculationClient(srv.URL, testSecret)

				err := cmd.call(client)

				assertErrContains(t, err, st.wantErr)
				assertRequest(t, requests, http.MethodPost, cmd.path)
			})
		}
	}
}

func TestCirculationClient_ServiceUnreachable(t *testing.T) {
	srv, _ := newFakeServer(t, http.StatusOK, "")
	srv.Close()
	client := NewCirculationClient(srv.URL, testSecret)
	ctx := context.Background()

	_, findErr := client.FindLoansNeedingNotice(ctx)
	assertErrContains(t, findErr, "calling circulation-service")
	assertErrContains(t, client.MarkNotified(ctx, "loan-1"), "calling circulation-service")
	assertErrContains(t, client.AccrueFines(ctx), "calling circulation-service")
}
