package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"collabotask/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

func newCSRFRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CSRF())
	r.GET("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.PUT("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.PATCH("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.DELETE("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/auth/logout", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCSRF_POST_WithHeader_Passes(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set("X-CSRF-Protection", "1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCSRF_POST_WithoutHeader_Returns403(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRF_PUT_WithoutHeader_Returns403(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/resource", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRF_PATCH_WithoutHeader_Returns403(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/resource", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRF_DELETE_WithoutHeader_Returns403(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}
}

func TestCSRF_GET_WithoutHeader_Passes(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for safe method GET, got %d", w.Code)
	}
}

func TestCSRF_AuthLogin_WithoutHeader_Returns403(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 (CSRF applies to /auth/* too), got %d", w.Code)
	}
}

func TestCSRF_AuthLogout_WithHeader_Passes(t *testing.T) {
	r := newCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("X-CSRF-Protection", "1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
