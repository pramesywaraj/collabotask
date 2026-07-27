package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

func newCORSRouter(cfg *config.CORSConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS(cfg))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.OPTIONS("/test", func(c *gin.Context) {})
	return r
}

func TestCORS_AllowlistedOriginReflected(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Protection"},
		AllowCredentials: true,
	}
	r := newCORSRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Expected ACAO http://localhost:3000, got %q", got)
	}
}

func TestCORS_NonAllowlistedOriginOmitted(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}
	r := newCORSRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no ACAO header for unlisted origin, got %q", got)
	}
}

func TestCORS_NoOriginHeader_NoACAO(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET"},
	}
	r := newCORSRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no ACAO header when no Origin sent, got %q", got)
	}
}

func TestCORS_Preflight_AdvertisesPatchAndCSRFHeader(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Protection"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	r := newCORSRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", w.Code)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if !headerContains(methods, "PATCH") {
		t.Errorf("Preflight should advertise PATCH, got: %s", methods)
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if !headerContains(headers, "X-CSRF-Protection") {
		t.Errorf("Preflight should advertise X-CSRF-Protection, got: %s", headers)
	}
}

func TestCORS_VaryOriginAlwaysPresent(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET"},
	}
	r := newCORSRouter(cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Vary"); got == "" {
		t.Error("Expected Vary: Origin header to be present")
	}
}

func headerContains(header, needle string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(part) == needle {
			return true
		}
	}
	return false
}
