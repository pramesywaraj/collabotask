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
	BoardHandler     *handler.BoardHandler
	ColumnHandler    *handler.ColumnHandler
	CardHandler      *handler.CardHandler
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
		workspaces.GET("/:workspace_id", cfg.WorkspaceHandler.GetWorkspaceDetail)
		workspaces.POST("/:workspace_id/member/invite", cfg.WorkspaceHandler.InviteMember)
		workspaces.DELETE("/:workspace_id/member/remove/:user_id", cfg.WorkspaceHandler.RemoveMember)

		boards := workspaces.Group("/:workspace_id/board")
		{
			boards.POST("", cfg.BoardHandler.CreateBoard)
			boards.GET("", cfg.BoardHandler.FetchListBoardsInWorkspace)
			boards.GET("/:board_id", cfg.BoardHandler.FetchBoardDetail)
			boards.GET("/:board_id/kanban", cfg.BoardHandler.FetchBoardKanban)
			boards.PATCH("/:board_id", cfg.BoardHandler.UpdateBoard)
			boards.POST("/:board_id/archive", cfg.BoardHandler.SetBoardArchivedStatus)
			boards.POST("/:board_id/invite", cfg.BoardHandler.InviteMembersToBoard)
			boards.DELETE("/:board_id/member", cfg.BoardHandler.RemoveMemberFromBoard)
			boards.GET("/:board_id/invitees", cfg.BoardHandler.FetchWorkspaceInviteesForBoard)
			boards.POST("/:board_id/join", cfg.BoardHandler.SelfJoinToBoard)
			boards.POST("/:board_id/leave", cfg.BoardHandler.LeaveBoard)
		}

		columns := boards.Group("/:board_id/columns")
		{
			columns.POST("", cfg.ColumnHandler.CreateColumn)
			columns.PATCH("/:column_id", cfg.ColumnHandler.UpdateColumn)
			columns.DELETE("/:column_id", cfg.ColumnHandler.DeleteColumn)
			columns.PATCH("/:column_id/position", cfg.ColumnHandler.UpdateColumnPosition)
		}

		cards := columns.Group("/:column_id/cards")
		{
			cards.POST("", cfg.CardHandler.CreateCard)
			cards.PATCH("/:card_id", cfg.CardHandler.UpdateCard)
			cards.DELETE("/:card_id", cfg.CardHandler.DeleteCard)
			cards.POST("/:card_id/move", cfg.CardHandler.MoveCardPosition)
		}
	}

	return routes
}
