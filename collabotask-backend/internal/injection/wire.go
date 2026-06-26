//go:build wireinject
// +build wireinject

package injection

import (
	"collabotask/internal/config"
	"collabotask/internal/infrastructure/database"
	"collabotask/internal/server"
	"collabotask/pkg/logger"

	"github.com/google/wire"
)

type App struct {
	Server  *server.Server
	DB      *database.DB
	Config  *config.Config
	Logger  *logger.Logger
	Cleanup func()
}

var (
	ConfigSet     = wire.NewSet(ProvideConfig)
	LoggerSet     = wire.NewSet(ProvideLogger)
	DBSet         = wire.NewSet(ProvideDB, ProvideCleanup)
	RepositorySet = wire.NewSet(
		ProvideUserRepository,
		ProvideWorkspaceRepository,
		ProvideWorkspaceMemberRepository,
		ProvideBoardRepository,
		ProvideBoardMemberRepository,
		ProvideColumnRepository,
		ProvideCardRepository,
	)
	UseCaseSet = wire.NewSet(
		ProvideAuthUseCase,
		ProvideWorkspaceUseCase,
		ProvideBoardUseCase,
		ProvideBoardAccessChecker,
		ProvideColumnUseCase,
		ProvideCardUseCase,
	)
	HandlerSet = wire.NewSet(
		ProvideAuthHandler,
		ProvideUserHandler,
		ProvideWorkspaceHandler,
		ProvideBoardHandler,
		ProvideColumnHandler,
		ProvideCardHandler,
	)
	RouterSet = wire.NewSet(ProvideRouter)
	ServerSet = wire.NewSet(ProvideServer)
)

func InitializeApp() (*App, error) {
	wire.Build(
		ConfigSet,
		LoggerSet,
		DBSet,
		RepositorySet,
		UseCaseSet,
		HandlerSet,
		RouterSet,
		ServerSet,
		wire.Struct(new(App), "*"),
	)

	return nil, nil
}
