package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/middleware"
	infraauth "collabotask/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func mustGenerateToken(t *testing.T, cfg *config.AuthConfig, userID uuid.UUID) string {
	t.Helper()
	cfgWithExpiry := *cfg
	cfgWithExpiry.JWTExpiration = time.Hour
	token, err := infraauth.GenerateToken(&cfgWithExpiry, userID, "user")
	if err != nil {
		t.Fatalf("mustGenerateToken: %v", err)
	}
	return token
}

func authCfg() *config.AuthConfig {
	return &config.AuthConfig{
		JWTSecret:    "test-secret-key-long-enough-32ch",
		CookieName:   "token",
		CookieSecure: false,
	}
}

func newAuthRouter(cfg *config.AuthConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", middleware.Auth(cfg), func(c *gin.Context) {
		userID, ok := middleware.GetUserID(c)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, userID.String())
	})
	return r
}

func TestAuth_NoCookie_Returns401(t *testing.T) {
	r := newAuthRouter(authCfg())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	r := newAuthRouter(authCfg())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "not.a.valid.jwt"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuth_ValidCookie_SetsUserIDInContext(t *testing.T) {
	cfg := authCfg()
	r := newAuthRouter(cfg)

	// Generate a real token using the same infrastructure the app uses.
	token := mustGenerateToken(t, cfg, uuid.New())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: cfg.CookieName, Value: token})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestAuth_BearerHeader_Rejected(t *testing.T) {
	cfg := authCfg()
	r := newAuthRouter(cfg)

	token := mustGenerateToken(t, cfg, uuid.New())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	// Cookie-only auth — Bearer header must not be accepted anymore.
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 (Bearer no longer accepted), got %d", w.Code)
	}
}
