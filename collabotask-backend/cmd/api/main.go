package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"collabotask/internal/adapter/http/handler"
	"collabotask/internal/adapter/http/router"
	"collabotask/internal/adapter/repository/postgres"
	"collabotask/internal/config"
	"collabotask/internal/infrastructure/database"
	"collabotask/internal/server"
	"collabotask/internal/usecase/auth"
	"collabotask/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatal("failed to connect to database: " + err.Error())
	}

	defer db.Close()

	if err := database.RunMigrations(cfg); err != nil {
		log.Fatal("failed to run migrations: " + err.Error())
	}

	userRepo := postgres.NewUserRepository(db.Pool)
	authUseCase := auth.NewAuthUseCase(userRepo, &cfg.Auth)
	authHandler := handler.NewAuthHandler(authUseCase)
	userHandler := handler.NewUserHandler(authUseCase)

	r := router.New(router.Config{
		Cfg:         cfg,
		Log:         log,
		AuthHandler: authHandler,
		UserHandler: userHandler,
	})

	srv := server.New(cfg, r)

	go func() {
		log.Info("🚀 Server starting on " + cfg.Server.Host + ":" + cfg.Server.Port)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server errors: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), srv.ShutdownTimeout())
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: " + err.Error())
	}

	log.Info("server exited")
}
