package router

import (
	"collabotask/internal/adapter/http/handler"
	"collabotask/internal/adapter/http/middleware"
	"collabotask/internal/config"
	"collabotask/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Cfg              *config.Config
	Log              *logger.Logger
	AuthHandler      *handler.AuthHandler
	UserHandler      *handler.UserHandler
	WorkspaceHandler *handler.WorkspaceHandler
}

func New(cfg Config) *gin.Engine {
	routes := gin.New()

	routes.Use(middleware.Recover(cfg.Log))
	routes.Use(middleware.Logger(cfg.Log))
	routes.Use(middleware.CORS(&cfg.Cfg.CORS))

	// Public routess
	auth := routes.Group("/auth")
	{
		auth.POST("/register", cfg.AuthHandler.Register)
		auth.POST("/login", cfg.AuthHandler.Login)
	}

	// Protected routess
	user := routes.Group("/user")
	user.Use(middleware.Auth(&cfg.Cfg.Auth))
	{
		user.GET("/profile", cfg.UserHandler.GetProfile)
	}

	workspaces := routes.Group("/workspace")
	workspaces.Use(middleware.Auth(&cfg.Cfg.Auth))
	{
		workspaces.POST("", cfg.WorkspaceHandler.CreateWorkspace)
		workspaces.GET("", cfg.WorkspaceHandler.ListWorkspaces)
		workspaces.POST("/:workspace_id/member/invite", cfg.WorkspaceHandler.InviteMember)
		workspaces.DELETE("/:workspace_id/member/remove/:user_id", cfg.WorkspaceHandler.RemoveMember)
	}

	return routes
}
