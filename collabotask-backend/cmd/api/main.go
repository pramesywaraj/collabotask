package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"collabotask/internal/infrastructure/database"
	"collabotask/internal/injection"
)

func main() {
	app, err := injection.InitializeApp()
	if err != nil {
		panic("failed to initialize app: " + err.Error())
	}

	cfg := app.Config
	log := app.Logger
	srv := app.Server

	if err := database.RunMigrations(cfg); err != nil {
		log.Fatal("failed to run migrations: " + err.Error())
	}

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
		app.Cleanup()
		log.Fatal("server forced to shutdown: " + err.Error())
	}

	log.Info("server exited")
}
