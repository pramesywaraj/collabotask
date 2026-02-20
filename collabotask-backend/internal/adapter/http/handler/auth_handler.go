package handler

import (
	"collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/usecase/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase auth.AuthUseCase
}

func NewAuthHandler(authUseCase auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	out, err := h.authUseCase.Register(c.Request.Context(), auth.RegisterInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		if err.Error() == "email already exists" {
			response.GenerateErrorResponse(c, errors.NewAppError(http.StatusConflict, errors.ErrCodeConflict, err.Error()))
			return
		}

		response.GenerateErrorResponse(c, errors.NewAppError(http.StatusConflict, errors.ErrCodeValidation, err.Error()))
		return
	}

	response.GenerateSuccessResponse(c, "User registered successfully", response.AuthResponse{
		User:  userDTOToResponse(out.User),
		Token: out.Token,
	}, http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	out, err := h.authUseCase.Login(c.Request.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.GenerateErrorResponse(c, errors.NewAppError(http.StatusUnauthorized, errors.ErrCodeUnauthorized, err.Error()))
		return
	}

	response.GenerateSuccessResponse(c, "Successfully logged in", response.AuthResponse{
		User:  userDTOToResponse(out.User),
		Token: out.Token,
	})
}

func userDTOToResponse(u auth.UserDTO) response.UserResponse {
	return response.UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		AvatarURL:  u.AvatarURL,
		SystemRole: u.SystemRole,
	}
}
