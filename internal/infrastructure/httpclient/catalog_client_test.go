package httpclient

import (
	"context"
	"net/http"
	"testing"
)

func TestCatalogClient_Reindex(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr string
	}{
		{name: "ok", status: http.StatusOK},
		{name: "not found", status: http.StatusNotFound, wantErr: "returned 404"},
		{name: "server error", status: http.StatusInternalServerError, wantErr: "returned 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, requests := newFakeServer(t, tt.status, "")
			client := NewCatalogClient(srv.URL, testSecret)

			err := client.Reindex(context.Background())

			assertErrContains(t, err, tt.wantErr)
			assertRequest(t, requests, http.MethodPost, "/api/v1/books/reindex")
		})
	}
}

func TestCatalogClient_ServiceUnreachable(t *testing.T) {
	srv, _ := newFakeServer(t, http.StatusOK, "")
	srv.Close()
	client := NewCatalogClient(srv.URL, testSecret)

	assertErrContains(t, client.Reindex(context.Background()), "calling catalog-service")
}
