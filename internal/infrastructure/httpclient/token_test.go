package httpclient

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestMintServiceToken(t *testing.T) {
	raw, err := mintServiceToken(testSecret, "lms-worker")
	if err != nil {
		t.Fatalf("mintServiceToken() error = %v", err)
	}

	claims := parseServiceToken(t, raw)
	if claims["sub"] != "lms-worker" {
		t.Errorf("sub = %v, want lms-worker", claims["sub"])
	}
	iat, iatErr := claims.GetIssuedAt()
	exp, expErr := claims.GetExpirationTime()
	if iatErr != nil || expErr != nil || iat == nil || exp == nil {
		t.Fatalf("iat = %v (%v), exp = %v (%v)", iat, iatErr, exp, expErr)
	}
	if got := exp.Sub(iat.Time).Seconds(); got != 60 {
		t.Errorf("token lifetime = %vs, want 60s", got)
	}
}

func TestMintServiceToken_RejectedWithAnotherSecret(t *testing.T) {
	raw, err := mintServiceToken(testSecret, "lms-worker")
	if err != nil {
		t.Fatalf("mintServiceToken() error = %v", err)
	}

	_, err = jwt.Parse(raw, func(*jwt.Token) (any, error) {
		return []byte("another-secret"), nil
	})
	if err == nil {
		t.Fatal("token validated with another secret, want an error")
	}
}
