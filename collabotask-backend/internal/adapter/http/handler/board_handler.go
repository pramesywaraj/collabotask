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

func (bh *BoardHandler) FetchBoardDetail(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	_, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.BoardDetailInput{
		RequesterID: userID,
		BoardID:     boardID,
	}

	out, err := bh.boardUseCase.BoardDetail(ctx.Request.Context(), input)
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
		"Board detail successfully fetched",
		response.BoardDetailDTOToResponse(out.Board),
	)
}

func (bh *BoardHandler) FetchListBoardsInWorkspace(ctx *gin.Context) {
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

	input := board.ListBoardsInput{
		WorkspaceID: workspaceID,
		RequesterID: userID,
	}

	out, err := bh.boardUseCase.ListBoardsInWorkspace(ctx.Request.Context(), input)
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
		"List boards in workspace successfully fetched",
		boards,
	)
}

func (bh *BoardHandler) FetchWorkspaceInviteesForBoard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	workspaceID, boardID, ok := parseBoardPathParams(ctx)
	if !ok {
		return
	}

	input := board.ListWorkspaceInviteesForBoardInput{
		RequesterID: userID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}

	out, err := bh.boardUseCase.ListWorkspaceInviteesForBoard(ctx.Request.Context(), input)
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
		"Workspace invitees for board successfully fetched",
		members,
	)
}

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
		"Board successfully updated",
		response.BoardDTOToResponse(out.Board),
	)
}

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
