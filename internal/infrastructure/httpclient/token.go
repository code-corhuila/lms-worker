// Package httpclient implements notice.CirculationClient and
// notice.CatalogClient over plain net/http — the same inter-service
// communication tool every other service in this system uses (no gRPC, no
// broker, AT-002).
package httpclient

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// mintServiceToken signs a short-lived internal token with
// INTERNAL_JWT_SECRET — deliberately not any end-user RS256 session token
// (rules/2-anexos/C-api-hexagonal.md, numeral 5.3.7). Shared by every client
// in this package.
func mintServiceToken(jwtSecret, subject string) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
