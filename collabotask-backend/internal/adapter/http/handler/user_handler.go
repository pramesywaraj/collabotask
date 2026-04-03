package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/middleware"
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
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, "Unauthorized"))
		return
	}

	user, err := h.authUseCase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
			return
		}
		response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		return
	}

	response.GenerateSuccessResponse(c, "Profile retrieved successfully", response.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: user.SystemRole,
	})
}
