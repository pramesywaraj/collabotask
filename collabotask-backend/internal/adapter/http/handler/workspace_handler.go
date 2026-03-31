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
