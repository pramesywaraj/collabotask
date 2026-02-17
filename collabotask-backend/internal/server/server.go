package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"collabotask/internal/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func New(cfg *config.Config, router *gin.Engine) *Server {
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, port)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  cfg.Server.Timeout,
			WriteTimeout: cfg.Server.Timeout,
			IdleTimeout:  2 * cfg.Server.Timeout,
		},
		shutdownTimeout: 10 * time.Second,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) ShutdownTimeout() time.Duration {
	return s.shutdownTimeout
}
