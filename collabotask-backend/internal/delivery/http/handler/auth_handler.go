package handler

import (
	"net/http"

	"collabotask/internal/config"
	"collabotask/internal/delivery/http/helper"
	"collabotask/internal/delivery/http/request"
	"collabotask/internal/delivery/http/response"
	"collabotask/internal/usecase/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase *auth.AuthUseCase
	cfg         *config.AuthConfig
}

func NewAuthHandler(authUseCase *auth.AuthUseCase, cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase, cfg: cfg}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.RegisterRequest true "Registration payload"
// @Param X-CSRF-Protection header string true "CSRF protection header (any non-empty value)"
// @Success 201 {object} response.AuthRegisterSuccessDoc "Created"
// @Failure 400 {object} response.Failure400BadRequestDoc "Validation error"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Missing CSRF header"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Security CSRF
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
		helper.HandleUseCaseError(ctx, err)
		return
	}

	helper.SetAuthCookie(ctx, ah.cfg, out.Token)
	response.GenerateSuccessResponse(ctx, "User registered successfully", response.AuthResponse{
		User: response.UserToResponse(out.User),
	}, http.StatusCreated)
}

// Login godoc
// @Summary Log in a user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "Login credentials"
// @Param X-CSRF-Protection header string true "CSRF protection header (any non-empty value)"
// @Success 200 {object} response.AuthLoginSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401LoginDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Missing CSRF header"
// @Security CSRF
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
		helper.HandleUseCaseError(ctx, err)
		return
	}

	helper.SetAuthCookie(ctx, ah.cfg, out.Token)
	response.GenerateSuccessResponse(ctx, "Successfully logged in", response.AuthResponse{
		User: response.UserToResponse(out.User),
	})
}

// Logout godoc
// @Summary Log out the current user
// @Tags auth
// @Param X-CSRF-Protection header string true "CSRF protection header (any non-empty value)"
// @Success 200 {object} response.SuccessDoc "OK"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Missing CSRF header"
// @Security CSRF
// @Router /auth/logout [post]
func (ah *AuthHandler) Logout(ctx *gin.Context) {
	helper.ClearAuthCookie(ctx, ah.cfg)
	response.GenerateSuccessResponse(ctx, "Successfully logged out", nil)
}
