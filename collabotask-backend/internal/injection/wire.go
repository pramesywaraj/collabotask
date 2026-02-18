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
	RepositorySet = wire.NewSet(ProvideUserRepository)
	UseCaseSet    = wire.NewSet(ProvideAuthUseCase)
	HandlerSet    = wire.NewSet(ProvideAuthHandler, ProvideUserHandler)
	RouterSet     = wire.NewSet(ProvideRouter)
	ServerSet     = wire.NewSet(ProvideServer)
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
