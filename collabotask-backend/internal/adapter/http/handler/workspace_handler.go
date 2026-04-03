package handler

//TODO:
// - Add update workspace handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/helper"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/workspace"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WorkspaceHandler struct {
	workspaceUseCase workspace.WorkspaceUseCase
}

func NewWorkspaceHandler(workspaceUseCase workspace.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceUseCase: workspaceUseCase,
	}
}

func workspaceDTOToResponse(d workspace.WorkspaceDTO) response.WorkspaceResponse {
	return response.WorkspaceResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		OwnerID:     d.OwnerID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func workspaceWithMetaDTOToResponse(d workspace.WorkspaceWithMetaDTO) response.WorkspaceWithMetaResponse {
	return response.WorkspaceWithMetaResponse{
		WorkspaceResponse: workspaceDTOToResponse(d.WorkspaceDTO),
		MemberCount:       d.MemberCount,
		BoardCount:        d.BoardCount,
		Role:              d.Role,
	}
}

func workspaceMemberDTOToResponse(d workspace.WorkspaceMemberDTO) response.WorkspaceMemberResponse {
	return response.WorkspaceMemberResponse{
		UserID:    d.UserID,
		Email:     d.Email,
		Name:      d.Name,
		AvatarURL: d.AvatarURL,
		Role:      d.Role,
		JoinedAt:  d.JoinedAt,
	}
}

func workspaceDetailDTOToResponse(d workspace.WorkspaceDetailDTO) response.WorkspaceDetailResponse {
	members := make([]response.WorkspaceMemberResponse, 0, len(d.Members))
	for _, member := range d.Members {
		members = append(members, workspaceMemberDTOToResponse(member))
	}

	return response.WorkspaceDetailResponse{
		WorkspaceResponse: workspaceDTOToResponse(d.WorkspaceDTO),
		UserRole:          d.UserRole,
		Members:           members,
	}
}

// CreateWorkspace godoc
// @Summary Create a new workspace for the authenticated user
// @Description Creates a workspace; returns 201 with workspace payload on success.
// @Tags workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.CreateWorkspaceRequest true "Create workspace payload"
// @Success 201 {object} response.WorkspaceCreateSuccessDoc "Created"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error (body or use case validation)"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid Bearer token"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace [post]
func (wh *WorkspaceHandler) CreateWorkspace(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	var req request.CreateWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := workspace.CreateWorkspaceInput{
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
	}

	out, err := wh.workspaceUseCase.CreateWorkspace(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Workspace created successfully",
		workspaceDTOToResponse(out.Workspace),
		http.StatusCreated,
	)
}

// ListWorkspaces godoc
// @Summary Get workspace list for authenticated user
// @Description Returns all workspaces the user belongs to.
// @Tags workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.WorkspaceListSuccessDoc "OK"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid Bearer token"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace [get]
func (wh *WorkspaceHandler) ListWorkspaces(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	out, err := wh.workspaceUseCase.ListWorkspaces(ctx.Request.Context(), workspace.ListWorkspacesInput{UserID: userID})
	if err != nil {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		return
	}

	workspaces := make([]response.WorkspaceWithMetaResponse, 0, len(out.Workspaces))
	for _, w := range out.Workspaces {
		workspaces = append(workspaces, workspaceWithMetaDTOToResponse(w))
	}

	response.GenerateSuccessResponse(
		ctx,
		"Workspace retrieved successfully",
		workspaces,
	)
}

// InviteMember godoc
// @Summary Invite users to a workspace by email
// @Description Requires workspace admin. Invalid workspace id or body validation returns 400.
// @Tags workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param body body request.InviteMemberRequest true "Email list"
// @Success 200 {object} response.WorkspaceInviteSuccessDoc "OK (message from use case; data may be null)"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation failed, invalid workspace id, or use case validation"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid Bearer token"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Requester is not workspace admin"
// @Failure 404 {object} response.Failure404NotFoundDoc "User not found"
// @Failure 409 {object} response.Failure409ConflictDoc "User already in workspace"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/member/invite [post]
func (wh *WorkspaceHandler) InviteMember(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid workspace id"),
		)
		return
	}

	var req request.InviteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := workspace.InviteMemberInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		Emails:      req.Emails,
	}

	out, err := wh.workspaceUseCase.InviteMember(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		switch {
		case errors.Is(err, domain.ErrNotWorkspaceAdmin):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
		case errors.Is(err, domain.ErrUserNotFound):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
		case errors.Is(err, domain.ErrAlreadyMember):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeConflict, err.Error()))
		default:
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		}
		return
	}

	response.GenerateSuccessResponse(ctx, out.Message, nil)
}

// RemoveMember godoc
// @Summary Remove a member from a workspace
// @Description Requires admin or appropriate permission; cannot remove yourself (400).
// @Tags workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param user_id path string true "Member user UUID to remove"
// @Success 200 {object} response.WorkspaceRemoveMemberSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id or invalid user id in path"
// @Failure 400 {object} response.Failure400ValidationDoc "Cannot remove yourself or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid Bearer token"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Not admin or user not in workspace"
// @Failure 404 {object} response.Failure404NotFoundDoc "Member not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/member/remove/{user_id} [delete]
func (wh *WorkspaceHandler) RemoveMember(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid workspace id"))
		return
	}

	memberUserID, ok := helper.ParseUUIDParams(ctx, "user_id")
	if !ok {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid user id"))
		return
	}

	err := wh.workspaceUseCase.RemoveMember(ctx.Request.Context(), workspace.RemoveMemberInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		UserID:      memberUserID,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotWorkspaceAdmin), errors.Is(err, domain.ErrUserNotInWorkspace):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
		case errors.Is(err, domain.ErrCannotRemoveYourself):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, err.Error()))
		case errors.Is(err, domain.ErrMemberNotFound):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
		default:
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		}
		return
	}

	response.GenerateSuccessResponse(ctx, "Member removed successfully", nil)
}

// GetWorkspaceDetail godoc
// @Summary Get workspace detail including members
// @Description Returns workspace metadata, current user role, and member list.
// @Tags workspace
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Success 200 {object} response.WorkspaceDetailSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id in path"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid Bearer token"
// @Failure 403 {object} response.Failure403ForbiddenDoc "User not in workspace"
// @Failure 404 {object} response.Failure404NotFoundDoc "Workspace or user not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id} [get]
func (wh *WorkspaceHandler) GetWorkspaceDetail(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid workspace id"),
		)
		return
	}

	input := workspace.WorkspaceDetailInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
	}

	out, err := wh.workspaceUseCase.WorkspaceDetail(
		ctx.Request.Context(),
		input,
	)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		switch {
		case errors.Is(err, domain.ErrUserNotInWorkspace):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
		case errors.Is(err, domain.ErrWorkspaceNotFound), errors.Is(err, domain.ErrUserNotFound):
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
		default:
			response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
		}
		return
	}

	response.GenerateSuccessResponse(ctx, "Workspace detail successfully fetched", workspaceDetailDTOToResponse(out.Workspace))
}
