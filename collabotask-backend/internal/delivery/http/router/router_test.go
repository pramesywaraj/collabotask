package router_test

import (
	"testing"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/handler"
	"collabotask/internal/delivery/http/router"
	"collabotask/pkg/logger"
)

// TestRouterBoots verifies that registering all routes doesn't cause Gin to panic
// due to wildcard conflicts (e.g. static /member/invite vs param /:user_id under /member/).
func TestRouterBoots(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("router.New panicked: %v", r)
		}
	}()

	router.New(router.Config{
		Cfg:              &config.Config{},
		Log:              logger.New(logger.Config{}),
		AuthHandler:      (*handler.AuthHandler)(nil),
		UserHandler:      (*handler.UserHandler)(nil),
		WorkspaceHandler: (*handler.WorkspaceHandler)(nil),
		BoardHandler:     (*handler.BoardHandler)(nil),
		ColumnHandler:    (*handler.ColumnHandler)(nil),
		CardHandler:      (*handler.CardHandler)(nil),
	})
}
