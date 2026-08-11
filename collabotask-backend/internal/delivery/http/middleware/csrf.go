package middleware

import (
	"net/http"

	apperrors "collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/response"

	"github.com/gin-gonic/gin"
)

// RFC 7231 safe methods — deliberately independent of the CORS method allowlist;
// do not consolidate. CSRF-safe = "non-mutating per RFC 7231" (GET/HEAD/OPTIONS);
// CORS-allowed = "browser may send cross-origin" — the two sets overlap by coincidence.
var csrfSafeMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// CSRF requires the X-CSRF-Protection header on all state-changing requests.
// Paired with SameSite=Lax cookies and an explicit CORS allowlist, this closes
// the cross-origin mutation vector for sibling-subdomain topologies.
//
// INVARIANT: do not extend CSRF checks to GET/HEAD/OPTIONS.
// The WebSocket handshake is a GET; the browser WebSocket API cannot send
// X-CSRF-Protection. Handshake CSWSH defense lives in OriginPatterns +
// SameSite=Lax, not in this middleware. See ws_handler.go.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if csrfSafeMethods[c.Request.Method] {
			c.Next()
			return
		}

		if c.GetHeader("X-CSRF-Protection") == "" {
			response.GenerateErrorResponse(c, apperrors.NewAppError(
				http.StatusForbidden,
				apperrors.ErrCodeForbidden,
				"X-CSRF-Protection header required",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
