package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/helper"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/board"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type BoardHandler struct {
	boardUseCase board.BoardUseCase
}

func NewBoardHandler(bu board.BoardUseCase) *BoardHandler {
	return &BoardHandler{
		boardUseCase: bu,
	}
}

func handleBoardError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotInWorkspace),
		errors.Is(err, domain.ErrBoardAccessDenied),
		errors.Is(err, domain.ErrBoardPermissionDenied),
		errors.Is(err, domain.ErrBoardOwnerCannotLeave):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
	case errors.Is(err, domain.ErrBoardNotFound),
		errors.Is(err, domain.ErrBoardMemberNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrMemberNotFound):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
	case errors.Is(err, domain.ErrConstraintViolation),
		errors.Is(err, domain.ErrAtLeastOneProvided),
		errors.Is(err, domain.ErrCannotRemoveYourself),
		errors.Is(err, domain.ErrBoardNoMembersToInvite):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, err.Error()))
	case errors.Is(err, domain.ErrBoardAlreadyMember),
		errors.Is(err, domain.ErrBoardCannotJoin):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeConflict, err.Error()))
	default:
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
	}
}

func parseBoardPathParams(ctx *gin.Context) (workspaceID, boardID uuid.UUID, ok bool) {
	workspaceID, okWorkspace := helper.ParseUUIDParams(ctx, "workspace_id")
	boardID, okBoard := helper.ParseUUIDParams(ctx, "board_id")

	if !okWorkspace || !okBoard {
		message := "Invalid or missing board id"
		if !okWorkspace {
			message = "Invalid or missing workspace id"
		}
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, message),
		)
		return uuid.Nil, uuid.Nil, false
	}

	return workspaceID, boardID, true
}

// CreateBoard godoc
// @Summary Create a board in a workspace
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param body body request.CreateBoardRequest true "Board payload"
// @Success 201 {object} response.BoardCreateSuccessDoc "Created"
// @Failure 400 {object} response.Failure400ValidationDoc "Invalid workspace id or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board [post]
func (bh *BoardHandler) CreateBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, ok := helper.ParseUUIDParams(ctx, "workspace_id")
	if !ok {
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(
				http.StatusBadRequest,
				apperrors.ErrCodeValidation,
				"Invalid or missing workspace id",
			),
		)
		return
	}

	var req request.CreateBoardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := board.CreateBoardInput{
		WorkspaceID:     workspaceID,
		RequesterID:     userID,
		Title:           req.Title,
		Description:     req.Description,
		BackgroundColor: req.BackgroundColor,
	}

	out, err := bh.boardUseCase.CreateBoard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Board created successfully",
		response.BoardDTOToResponse(out.Board),
		http.StatusCreated,
	)
}

// FetchBoardDetail godoc
// @Summary Get board detail
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Success 200 {object} response.BoardDetailSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid or missing workspace/board id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id} [get]
func (bh *BoardHandler) GetBoardDetail(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.GetBoardDetailInput{
		RequesterID: userID,
		BoardID:     boardID,
	}

	out, err := bh.boardUseCase.GetBoardDetail(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Board detail fetched successfully",
		response.BoardDetailDTOToResponse(out.Board),
	)
}

// FetchListBoardsInWorkspace godoc
// @Summary List boards in a workspace
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Success 200 {object} response.BoardListInWorkspaceSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid workspace id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board [get]
func (bh *BoardHandler) GetBoardsInWorkspace(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, okWorkspace := helper.ParseUUIDParams(ctx, "workspace_id")
	if !okWorkspace {
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(
				http.StatusBadRequest,
				apperrors.ErrCodeValidation,
				"Invalid or missing workspace id",
			),
		)
		return
	}

	input := board.GetBoardsInput{
		WorkspaceID: workspaceID,
		RequesterID: userID,
	}

	out, err := bh.boardUseCase.GetBoardsInWorkspace(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	boards := make([]response.BoardWithMetaResponse, 0, len(out.Boards))
	for _, b := range out.Boards {
		boards = append(boards, response.BoardWithMetaDTOToResponse(b))
	}

	response.GenerateSuccessResponse(
		ctx,
		"Boards in Workspace retrieved successfully",
		boards,
	)
}

// FetchWorkspaceInviteesForBoard godoc
// @Summary List workspace members eligible as board invitees
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Success 200 {object} response.BoardInviteesListSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid or missing workspace/board id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/invitees [get]
func (bh *BoardHandler) GetWorkspaceInviteesForBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.GetWorkspaceInviteesForBoardInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}

	out, err := bh.boardUseCase.GetWorkspaceInviteesForBoard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	members := make([]response.BoardInviteeResponse, 0, len(out.Members))
	for _, m := range out.Members {
		members = append(members, response.BoardInviteeDTOToResponse(m))
	}

	response.GenerateSuccessResponse(
		ctx,
		"Workspace invitees for board retrieved successfully",
		members,
	)
}

// UpdateBoard godoc
// @Summary Update board fields
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param body body request.UpdateBoardRequest true "Partial update"
// @Success 200 {object} response.BoardUpdateSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id} [patch]
func (bh *BoardHandler) UpdateBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	var req request.UpdateBoardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := board.UpdateBoardInput{
		RequesterID:     userID,
		BoardID:         boardID,
		Title:           req.Title,
		BackgroundColor: req.BackgroundColor,
	}
	if req.Description.Present {
		input.DescriptionPresent = true
		input.Description = req.Description.Value
	}

	out, err := bh.boardUseCase.UpdateBoard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Board updated successfully",
		response.BoardDTOToResponse(out.Board),
	)
}

// SetBoardArchivedStatus godoc
// @Summary Set board archived flag
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param body body request.SetArchivedBoardRequest true "Archived flag"
// @Success 200 {object} response.BoardArchiveSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/archive [post]
func (bh *BoardHandler) SetBoardArchivedStatus(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	var req request.SetArchivedBoardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := board.SetArchivedInput{
		RequesterID: userID,
		BoardID:     boardID,
		IsArchived:  req.IsArchived,
	}

	out, err := bh.boardUseCase.SetArchived(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Board archived status changed",
		response.BoardDTOToResponse(out.Board),
	)
}

// InviteMembersToBoard godoc
// @Summary Invite workspace members to the board by user id
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param body body request.InviteMemberBoardRequest true "User IDs to invite"
// @Success 200 {object} response.BoardInviteMembersSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/invite [post]
func (bh *BoardHandler) InviteMembersToBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	var req request.InviteMemberBoardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := board.InviteMemberInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserIDs:     req.UserIDs,
	}

	err := bh.boardUseCase.InviteMember(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Members have been invited to the board",
		nil,
	)
}

// RemoveMemberFromBoard godoc
// @Summary Remove a member from the board
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param body body request.RemoveMemberBoardRequest true "Member user id"
// @Success 200 {object} response.BoardRemoveMemberSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/member [delete]
func (bh *BoardHandler) RemoveMemberFromBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	var req request.RemoveMemberBoardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := board.RemoveMemberInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserID:      req.UserID,
	}

	err := bh.boardUseCase.RemoveMember(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Member has been removed from the board",
		nil,
	)
}

// SelfJoinToBoard godoc
// @Summary Join the board (self-service)
// @Description For now, its only eligible for Workspace Admin
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Success 200 {object} response.BoardSelfJoinSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid ids"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 409 {object} response.Failure409ConflictDoc "Already a member or cannot join"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/join [post]
func (bh *BoardHandler) SelfJoinToBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.SelfJoinBoardInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}

	err := bh.boardUseCase.SelfJoinBoard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Successfully joined the board",
		nil,
	)
}

// LeaveBoard godoc
// @Summary Leave the board
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Success 200 {object} response.BoardLeaveSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid ids"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden (e.g. owner cannot leave)"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/leave [post]
func (bh *BoardHandler) LeaveBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.LeaveBoardInput{
		RequesterID: userID,
		BoardID:     boardID,
	}

	err := bh.boardUseCase.LeaveBoard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Successfully left the board",
		nil,
	)
}

// FetchBoardKanban godoc
// @Summary Get board kanban (columns and cards)
// @Tags board
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Success 200 {object} response.BoardKanbanSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid or missing workspace/board id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/kanban [get]
func (bh *BoardHandler) GetBoardKanban(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.GetBoardKanbanInput{
		RequesterID: userID,
		BoardID:     boardID,
	}

	out, err := bh.boardUseCase.GetBoardKanban(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleBoardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Board kanban retrieved successfully",
		response.BoardKanbanToResponse(out.Columns),
	)
}
