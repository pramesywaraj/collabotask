package helper_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/helper"

	"github.com/gin-gonic/gin"
)

func ginCtxWithRecorder() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	return c, w
}

func TestAuthCookieSet_Secure(t *testing.T) {
	cfg := &config.AuthConfig{
		CookieSecure:  true,
		CookieName:    "__Host-token",
		JWTExpiration: 24 * 60 * 60 * 1e9, // 24h in nanoseconds
	}

	c, w := ginCtxWithRecorder()
	helper.SetAuthCookie(c, cfg, "test.jwt.token")

	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("Expected Set-Cookie header, got none")
	}

	assertCookieAttr(t, cookie, "HttpOnly")
	assertCookieAttr(t, cookie, "SameSite=Lax")
	assertCookieAttr(t, cookie, "Path=/")
	assertCookieAttr(t, cookie, "Secure")
	assertCookieAttr(t, cookie, "__Host-token=")

	if strings.Contains(cookie, "Domain=") {
		t.Error("__Host- cookie must not set Domain")
	}
}

func TestAuthCookieSet_Insecure(t *testing.T) {
	cfg := &config.AuthConfig{
		CookieSecure:  false,
		CookieName:    "token",
		JWTExpiration: 24 * 60 * 60 * 1e9,
	}

	c, w := ginCtxWithRecorder()
	helper.SetAuthCookie(c, cfg, "test.jwt.token")

	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("Expected Set-Cookie header, got none")
	}

	assertCookieAttr(t, cookie, "HttpOnly")
	assertCookieAttr(t, cookie, "SameSite=Lax")
	assertCookieAttr(t, cookie, "Path=/")
	assertCookieAttr(t, cookie, "token=")

	if strings.Contains(cookie, "; Secure") {
		t.Error("Insecure cookie must not set Secure flag")
	}
}

func TestAuthCookieClear(t *testing.T) {
	cfg := &config.AuthConfig{
		CookieSecure: true,
		CookieName:   "__Host-token",
	}

	c, w := ginCtxWithRecorder()
	helper.ClearAuthCookie(c, cfg)

	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("Expected Set-Cookie header from Clear, got none")
	}

	if !strings.Contains(cookie, "Max-Age=0") {
		t.Errorf("Clear must set Max-Age=0, got: %s", cookie)
	}
	assertCookieAttr(t, cookie, "Path=/")
	assertCookieAttr(t, cookie, "Secure")
}

func assertCookieAttr(t *testing.T, cookie, attr string) {
	t.Helper()
	if !strings.Contains(cookie, attr) {
		t.Errorf("Expected cookie to contain %q, got: %s", attr, cookie)
	}
}
