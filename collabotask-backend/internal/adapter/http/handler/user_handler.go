package handler

import (
	"collabotask/internal/adapter/http/middleware"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/usecase/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authUseCase auth.AuthUseCase
}

func NewUserHandler(authUseCase auth.AuthUseCase) *UserHandler {
	return &UserHandler{authUseCase: authUseCase}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Message: "unauthorized"})
		return
	}

	user, err := h.authUseCase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: user.SystemRole,
	})
}
