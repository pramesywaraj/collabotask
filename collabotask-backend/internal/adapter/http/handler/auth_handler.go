package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase auth.AuthUseCase
}

func NewAuthHandler(authUseCase auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.RegisterRequest true "Registration payload"
// @Success 201 {object} response.AuthRegisterSuccessDoc "Created"
// @Failure 400 {object} response.Failure400BadRequestDoc "Validation error"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Router /auth/register [post]
func (ah *AuthHandler) Register(ctx *gin.Context) {
	var req request.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	out, err := ah.authUseCase.Register(ctx.Request.Context(), auth.RegisterInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeConflict, err.Error()))
			return
		}

		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeValidation, err.Error()))
		return
	}

	response.GenerateSuccessResponse(ctx, "User registered successfully", response.AuthResponse{
		User:  response.UserDTOToResponse(out.User),
		Token: out.Token,
	}, http.StatusCreated)
}

// Login godoc
// @Summary Log in a user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "Login credentials"
// @Success 200 {object} response.AuthLoginSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401LoginDoc "Unauthorized"
// @Router /auth/login [post]
func (ah *AuthHandler) Login(ctx *gin.Context) {
	var req request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	out, err := ah.authUseCase.Login(ctx.Request.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, err.Error()))
		return
	}

	response.GenerateSuccessResponse(ctx, "Successfully logged in", response.AuthResponse{
		User:  response.UserDTOToResponse(out.User),
		Token: out.Token,
	})
}
