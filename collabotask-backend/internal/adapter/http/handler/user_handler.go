package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/helper"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	authUseCase auth.AuthUseCase
}

func NewUserHandler(authUseCase auth.AuthUseCase) *UserHandler {
	return &UserHandler{authUseCase: authUseCase}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile. Requires a valid Bearer JWT.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
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
		if errors.Is(err, domain.ErrUserNotFound) {
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
			return
		}
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		return
	}

	response.GenerateSuccessResponse(ctx, "Profile retrieved successfully", response.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: user.SystemRole,
	})
}
