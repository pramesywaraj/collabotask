package middleware

import (
	"net/http"

	apperrors "collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/response"

	"github.com/gin-gonic/gin"
)

var csrfSafeMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// CSRF requires the X-CSRF-Protection header on all state-changing requests.
// Paired with SameSite=Lax cookies and an explicit CORS allowlist, this closes
// the cross-origin mutation vector for sibling-subdomain topologies.
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
