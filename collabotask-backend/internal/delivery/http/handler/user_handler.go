package handler

import (
	"collabotask/internal/delivery/http/helper"
	"collabotask/internal/delivery/http/response"
	"collabotask/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authUseCase *auth.AuthUseCase
}

func NewUserHandler(authUseCase *auth.AuthUseCase) *UserHandler {
	return &UserHandler{authUseCase: authUseCase}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile. Requires a valid Bearer JWT.
// @Tags user
// @Accept json
// @Produce json
// @Security CSRF
// @Success 200 {object} response.UserProfileSuccessDoc "OK"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized (missing/invalid token or user context)"
// @Failure 404 {object} response.Failure404NotFoundDoc "User not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	user, err := h.authUseCase.GetProfile(ctx.Request.Context(), userID)
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Profile retrieved successfully", response.UserToResponse(*user))
}
