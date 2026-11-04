package httpclient

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-internal-secret"

type recordedRequest struct {
	method string
	path   string
	auth   string
}

// newFakeServer answers every request with status and body, and records what
// it received so the test can assert on it from its own goroutine.
func newFakeServer(t *testing.T, status int, body string) (*httptest.Server, <-chan recordedRequest) {
	t.Helper()
	requests := make(chan recordedRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- recordedRequest{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization")}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, requests
}

// assertRequest checks the route and that the call carried a valid service token.
func assertRequest(t *testing.T, requests <-chan recordedRequest, method, path string) {
	t.Helper()
	got := <-requests
	if got.method != method || got.path != path {
		t.Errorf("request = %s %s, want %s %s", got.method, got.path, method, path)
	}
	claims := parseServiceToken(t, strings.TrimPrefix(got.auth, "Bearer "))
	if claims["sub"] != "lms-worker" {
		t.Errorf("sub = %v, want lms-worker", claims["sub"])
	}
}

func parseServiceToken(t *testing.T, raw string) jwt.MapClaims {
	t.Helper()
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return []byte(testSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		t.Fatalf("parsing service token: %v", err)
	}
	return claims
}

func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("error = %v, want nil", err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want it to contain %q", err, want)
	}
}
