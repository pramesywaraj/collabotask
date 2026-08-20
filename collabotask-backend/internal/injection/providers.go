package injection

import (
	"github.com/gin-gonic/gin"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/handler"
	"collabotask/internal/delivery/http/router"
	"collabotask/internal/domain/repository"
	infraauth "collabotask/internal/infrastructure/auth"
	"collabotask/internal/infrastructure/database"
	"collabotask/internal/realtime"
	"collabotask/internal/realtime/broadcast"
	"collabotask/internal/repository/postgres"
	"collabotask/internal/server"
	"collabotask/internal/usecase/auth"
	"collabotask/internal/usecase/board"
	"collabotask/internal/usecase/card"
	"collabotask/internal/usecase/column"
	"collabotask/internal/usecase/common"
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
func ProvideBoardRepository(db *database.DB) repository.BoardRepository {
	return postgres.NewBoardRepository(db.Pool)
}
func ProvideBoardMemberRepository(db *database.DB) repository.BoardMemberRepository {
	return postgres.NewBoardMemberRepository(db.Pool)
}
func ProvideColumnRepository(db *database.DB) repository.ColumnRepository {
	return postgres.NewColumnRepository(db.Pool)
}
func ProvideCardRepository(db *database.DB) repository.CardRepository {
	return postgres.NewCardRepository(db.Pool)
}
func ProvideActivityRepository(db *database.DB) repository.ActivityRepository {
	return postgres.NewActivityRepository(db.Pool)
}

// UseCase
func ProvideAuthUseCase(userRepo repository.UserRepository, cfg *config.Config) *auth.AuthUseCase {
	hasher := infraauth.NewBcryptHasher(&cfg.Auth)
	tokens := infraauth.NewJWTGenerator(&cfg.Auth)
	return auth.NewAuthUseCase(userRepo, hasher, tokens)
}
func ProvideWorkspaceUseCase(
	workspaceRepo repository.WorkspaceRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
	activityRepo repository.ActivityRepository,
) *workspace.WorkspaceUseCase {
	return workspace.NewWorkspaceUseCase(workspaceRepo, workspaceMemberRepo, userRepo, activityRepo)
}
func ProvideBroadcaster(hub *realtime.Hub) common.Broadcaster {
	return broadcast.NewHubBroadcaster(hub)
}
func ProvideBoardUseCase(
	boardAccessChecker common.BoardAccessChecker,
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
	columnRepo repository.ColumnRepository,
	cardRepo repository.CardRepository,
	activityRepo repository.ActivityRepository,
	broadcaster common.Broadcaster,
) *board.BoardUseCase {
	return board.NewBoardUseCase(boardAccessChecker, boardRepo, boardMemberRepo, workspaceMemberRepo, userRepo, columnRepo, cardRepo, activityRepo, broadcaster)
}
func ProvideColumnUseCase(
	columnRepo repository.ColumnRepository,
	boardAccessChecker common.BoardAccessChecker,
	activityRepo repository.ActivityRepository,
	broadcaster common.Broadcaster,
) *column.ColumnUseCase {
	return column.NewColumnUseCase(columnRepo, boardAccessChecker, activityRepo, broadcaster)
}
func ProvideCardUseCase(
	cardRepo repository.CardRepository,
	columnRepo repository.ColumnRepository,
	userRepo repository.UserRepository,
	boardAccessChecker common.BoardAccessChecker,
	boardMemberRepo repository.BoardMemberRepository,
	activityRepo repository.ActivityRepository,
	broadcaster common.Broadcaster,
) *card.CardUseCase {
	return card.NewCardUseCase(cardRepo, columnRepo, userRepo, boardAccessChecker, boardMemberRepo, activityRepo, broadcaster)
}

// Common use cases
func ProvideBoardAccessChecker(
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
) common.BoardAccessChecker {
	return common.NewBoardAccessChecker(boardRepo, boardMemberRepo, workspaceMemberRepo)
}

// Handler
func ProvideAuthHandler(authUseCase *auth.AuthUseCase, cfg *config.Config) *handler.AuthHandler {
	return handler.NewAuthHandler(authUseCase, &cfg.Auth)
}
func ProvideUserHandler(authUseCase *auth.AuthUseCase) *handler.UserHandler {
	return handler.NewUserHandler(authUseCase)
}
func ProvideWorkspaceHandler(workspaceUseCase *workspace.WorkspaceUseCase) *handler.WorkspaceHandler {
	return handler.NewWorkspaceHandler(workspaceUseCase)
}
func ProvideBoardHandler(boardUseCase *board.BoardUseCase) *handler.BoardHandler {
	return handler.NewBoardHandler(boardUseCase)
}
func ProvideColumnHandler(columnUseCase *column.ColumnUseCase) *handler.ColumnHandler {
	return handler.NewColumnHandler(columnUseCase)
}
func ProvideCardHandler(cardUseCase *card.CardUseCase) *handler.CardHandler {
	return handler.NewCardHandler(cardUseCase)
}

// ProvideHub is the single in-memory hub for the process (one registry, one lock).
// Its presence callback is wired by ProvideWSHandler via hub.SetPresence.
func ProvideHub() *realtime.Hub {
	return realtime.NewHub()
}

func ProvideWSHandler(
	cfg *config.Config,
	hub *realtime.Hub,
	access common.BoardAccessChecker,
) *handler.WSHandler {
	return handler.NewWSHandler(hub, access, cfg.WSOriginPatterns())
}

// Router
func ProvideRouter(
	cfg *config.Config,
	log *logger.Logger,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	workspaceHandler *handler.WorkspaceHandler,
	boardHandler *handler.BoardHandler,
	columnHandler *handler.ColumnHandler,
	cardHandler *handler.CardHandler,
	wsHandler *handler.WSHandler,
) *gin.Engine {
	return router.New(router.Config{
		Cfg:              cfg,
		Log:              log,
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		WorkspaceHandler: workspaceHandler,
		BoardHandler:     boardHandler,
		ColumnHandler:    columnHandler,
		CardHandler:      cardHandler,
		WSHandler:        wsHandler,
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
