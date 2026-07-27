package helper

import (
	"net/http"
	"time"

	"collabotask/internal/config"

	"github.com/gin-gonic/gin"
)

// SetAuthCookie writes the JWT as an httpOnly cookie with the configured attributes.
func SetAuthCookie(c *gin.Context, cfg *config.AuthConfig, jwt string) {
	maxAge := int(cfg.JWTExpiration / time.Second)
	http.SetCookie(c.Writer, buildCookie(cfg, jwt, maxAge))
}

// ClearAuthCookie expires the auth cookie by setting Max-Age=0 with matching attributes.
// In Go's http.Cookie, MaxAge=-1 renders as Max-Age=0 in the header.
func ClearAuthCookie(c *gin.Context, cfg *config.AuthConfig) {
	http.SetCookie(c.Writer, buildCookie(cfg, "", -1))
}

func buildCookie(cfg *config.AuthConfig, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.CookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}
