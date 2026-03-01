package injection

import (
	"github.com/gin-gonic/gin"

	"collabotask/internal/adapter/http/handler"
	"collabotask/internal/adapter/http/router"
	"collabotask/internal/adapter/repository/postgres"
	"collabotask/internal/config"
	"collabotask/internal/domain/repository"
	"collabotask/internal/infrastructure/database"
	"collabotask/internal/server"
	"collabotask/internal/usecase/auth"
	"collabotask/internal/usecase/workspace"
	"collabotask/pkg/logger"
)

func ProvideConfig() (*config.Config, error) {
	return config.Load()
}

func ProvideLogger(cfg *config.Config) *logger.Logger {
	return logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})
}

func ProvideDB(cfg *config.Config) (*database.DB, error) {
	return database.NewDB(cfg)
}

// Repository
func ProvideUserRepository(db *database.DB) repository.UserRepository {
	return postgres.NewUserRepository(db.Pool)
}

func ProvideWorkspaceRepository(db *database.DB) repository.WorkspaceRepository {
	return postgres.NewWorkspaceRepository(db.Pool)
}

func ProvideWorkspaceMemberRepository(db *database.DB) repository.WorkspaceMemberRepository {
	return postgres.NewWorkspaceMemberRepository(db.Pool)
}

// UseCase
func ProvideAuthUseCase(userRepo repository.UserRepository, cfg *config.Config) auth.AuthUseCase {
	return auth.NewAuthUseCase(userRepo, &cfg.Auth)
}

func ProvideWorkspaceUseCase(
	workspaceRepo repository.WorkspaceRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,

) workspace.WorkspaceUseCase {
	return workspace.NewWorkspaceUseCase(workspaceRepo, workspaceMemberRepo, userRepo)
}

// Handler
func ProvideAuthHandler(authUseCase auth.AuthUseCase) *handler.AuthHandler {
	return handler.NewAuthHandler(authUseCase)
}

func ProvideUserHandler(authUseCase auth.AuthUseCase) *handler.UserHandler {
	return handler.NewUserHandler(authUseCase)
}

func ProvideWorkspaceHandler(workspaceUseCase workspace.WorkspaceUseCase) *handler.WorkspaceHandler {
	return handler.NewWorkspaceHandler(workspaceUseCase)
}

// Router
func ProvideRouter(
	cfg *config.Config,
	log *logger.Logger,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	workspaceHandler *handler.WorkspaceHandler,
) *gin.Engine {
	return router.New(router.Config{
		Cfg:              cfg,
		Log:              log,
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		WorkspaceHandler: workspaceHandler,
	})
}

// Server
func ProvideServer(cfg *config.Config, r *gin.Engine) *server.Server {
	return server.New(cfg, r)
}

// Cleanup
func ProvideCleanup(db *database.DB) func() {
	return func() { db.Close() }
}
