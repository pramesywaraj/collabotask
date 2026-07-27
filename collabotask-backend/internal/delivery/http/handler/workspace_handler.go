package handler

//TODO:
// - Add update workspace handler

import (
	apperrors "collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/helper"
	"collabotask/internal/delivery/http/request"
	"collabotask/internal/delivery/http/response"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/workspace"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	workspaceUseCase *workspace.WorkspaceUseCase
}

func NewWorkspaceHandler(workspaceUseCase *workspace.WorkspaceUseCase) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceUseCase: workspaceUseCase,
	}
}

// CreateWorkspace godoc
// @Summary Create a new workspace for the authenticated user
// @Description Creates a workspace; returns 201 with workspace payload on success.
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Param body body request.CreateWorkspaceRequest true "Create workspace payload"
// @Success 201 {object} response.WorkspaceCreateSuccessDoc "Created"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error (body or use case validation)"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
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
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Workspace created successfully",
		response.WorkspaceToResponse(out.Workspace),
		http.StatusCreated,
	)
}

// ListWorkspaces godoc
// @Summary Get workspace list for authenticated user
// @Description Returns all workspaces the user belongs to.
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Success 200 {object} response.WorkspaceListSuccessDoc "OK"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace [get]
func (wh *WorkspaceHandler) GetWorkspaces(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	out, err := wh.workspaceUseCase.GetWorkspaces(ctx.Request.Context(), workspace.GetWorkspacesInput{UserID: userID})
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	workspaces := make([]response.WorkspaceWithMetaResponse, 0, len(out.Workspaces))
	for _, w := range out.Workspaces {
		workspaces = append(workspaces, response.WorkspaceWithMetaToResponse(w))
	}

	response.GenerateSuccessResponse(
		ctx,
		"Workspaces retrieved successfully",
		workspaces,
	)
}

// InviteMember godoc
// @Summary Invite users to a workspace by email
// @Description Requires workspace admin. Each email is processed independently; domain errors (NOT_FOUND, CONFLICT) are reported per-email in the results array, not as HTTP error codes. HTTP 200 is returned as long as the batch runs. Infrastructure errors or a non-admin requester still return 4xx/5xx.
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Param body body request.InviteMemberRequest true "Email list"
// @Success 200 {object} response.WorkspaceInviteSuccessDoc "OK — per-email results in data array"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation failed or invalid workspace id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Requester is not workspace admin"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error (infra failure mid-batch)"
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
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Invitation processed", response.InviteResultsToResponse(out.Results))
}

// RemoveMember godoc
// @Summary Remove a member from a workspace
// @Description Requires admin or appropriate permission; cannot remove yourself (400).
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Param user_id path string true "Member user UUID to remove"
// @Success 200 {object} response.WorkspaceRemoveMemberSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id or invalid user id in path"
// @Failure 400 {object} response.Failure400ValidationDoc "Cannot remove yourself or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
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
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Member removed successfully", nil)
}

// SetMemberRole godoc
// @Summary Promote or demote a workspace member's role
// @Description Requires workspace ADMIN. Idempotent: setting the existing role returns 200. Cannot demote owner or last admin.
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Param user_id path string true "Target member user UUID"
// @Param body body request.SetMemberRoleRequest true "New role"
// @Success 200 {object} response.WorkspaceSetMemberRoleSuccessDoc "OK — updated member"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Not admin, demoting owner, or insufficient authority"
// @Failure 404 {object} response.Failure404NotFoundDoc "Target member not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Last admin guard triggered"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/member/{user_id}/role [patch]
func (wh *WorkspaceHandler) SetMemberRole(ctx *gin.Context) {
	requesterID, ok := helper.GetAndCheckUserID(ctx)
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

	var req request.SetMemberRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	out, err := wh.workspaceUseCase.SetMemberRole(ctx.Request.Context(), workspace.SetMemberRoleInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		UserID:      memberUserID,
		Role:        entity.WorkspaceRole(req.Role),
	})
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Member role updated successfully", out.Member)
}

// LeaveWorkspace godoc
// @Summary Leave a workspace
// @Description Authenticated user leaves the workspace. Owner cannot leave; last admin cannot leave.
// @Tags workspace
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Success 200 {object} response.SuccessNullDataDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Owner cannot leave"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not a member"
// @Failure 409 {object} response.Failure409ConflictDoc "Last admin guard"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/leave [post]
func (wh *WorkspaceHandler) LeaveWorkspace(ctx *gin.Context) {
	requesterID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid workspace id"))
		return
	}

	err := wh.workspaceUseCase.LeaveWorkspace(ctx.Request.Context(), workspace.LeaveWorkspaceInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Left workspace successfully", nil)
}

// DeleteWorkspace godoc
// @Summary Delete a workspace
// @Description Hard delete; only the workspace owner may call this. DB cascades boards/columns/cards/members.
// @Tags workspace
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Success 200 {object} response.SuccessNullDataDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Not the workspace owner"
// @Failure 404 {object} response.Failure404NotFoundDoc "Workspace not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id} [delete]
func (wh *WorkspaceHandler) DeleteWorkspace(ctx *gin.Context) {
	requesterID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, "Invalid workspace id"))
		return
	}

	err := wh.workspaceUseCase.DeleteWorkspace(ctx.Request.Context(), workspace.DeleteWorkspaceInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Workspace deleted successfully", nil)
}

// GetWorkspaceDetail godoc
// @Summary Get workspace detail including members
// @Description Returns workspace metadata, current user role, and member list.
// @Tags workspace
// @Accept json
// @Produce json
// @Security CSRF
// @Param workspace_id path string true "Workspace UUID"
// @Success 200 {object} response.WorkspaceDetailSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id in path"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Missing or invalid authentication cookie"
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

	input := workspace.GetWorkspaceDetailInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
	}

	out, err := wh.workspaceUseCase.GetWorkspaceDetail(
		ctx.Request.Context(),
		input,
	)
	if err != nil {
		helper.HandleUseCaseError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(ctx, "Workspace detail retrieved successfully", response.WorkspaceDetailToResponse(out.Workspace))
}
