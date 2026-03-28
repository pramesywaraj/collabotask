package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/helper"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/column"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ColumnHandler struct {
	columnUseCase column.ColumnUseCase
}

func NewColumnHandler(cu column.ColumnUseCase) *ColumnHandler {
	return &ColumnHandler{
		columnUseCase: cu,
	}
}

func handleColumnError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotInWorkspace),
		errors.Is(err, domain.ErrBoardAccessDenied),
		errors.Is(err, domain.ErrBoardPermissionDenied),
		errors.Is(err, domain.ErrNotWorkspaceAdmin):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
	case errors.Is(err, domain.ErrBoardNotFound),
		errors.Is(err, domain.ErrColumnNotFound),
		errors.Is(err, domain.ErrWorkspaceNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrMemberNotFound),
		errors.Is(err, domain.ErrColumnNotInBoard):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
	case errors.Is(err, domain.ErrConstraintViolation),
		errors.Is(err, domain.ErrAtLeastOneProvided):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, err.Error()))
	case errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrBoardAlreadyMember),
		errors.Is(err, domain.ErrInconsistentState):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeConflict, err.Error()))
	default:
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
	}
}

func parseColumnPathParams(ctx *gin.Context) (boardID, columnID uuid.UUID, ok bool) {
	boardID, okBoard := helper.ParseUUIDParams(ctx, "board_id")
	columnID, okColumn := helper.ParseUUIDParams(ctx, "column_id")

	if !okBoard || !okColumn {
		var errMessage string
		switch {
		case !okBoard && !okColumn:
			errMessage = "Invalid or missing board and column ids"
		case !okBoard:
			errMessage = "Invalid or missing board id"
		default:
			errMessage = "Invalid or missing column id"
		}
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, errMessage),
		)
		return uuid.Nil, uuid.Nil, false
	}

	return boardID, columnID, true
}

func (ch *ColumnHandler) CreateColumn(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, ok := helper.ParseUUIDParams(ctx, "board_id")
	if !ok {
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(
				http.StatusBadRequest,
				apperrors.ErrCodeValidation,
				"Invalid or missing board id",
			),
		)
		return
	}

	var req request.CreateColumnRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := column.CreateColumnInput{
		BoardID:     boardID,
		Title:       req.Title,
		RequesterID: userID,
	}

	out, err := ch.columnUseCase.CreateColumn(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleColumnError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Column created successfully",
		response.ColumnDTOToResponse(out.Column),
		http.StatusCreated,
	)
}

func (ch *ColumnHandler) UpdateColumn(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, ok := parseColumnPathParams(ctx)
	if !ok {
		return
	}

	var req request.UpdateColumnRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := column.UpdateColumnInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       req.Title,
		RequesterID: userID,
	}

	out, err := ch.columnUseCase.UpdateColumn(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleColumnError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Column updated successfully",
		response.ColumnDTOToResponse(out.Column),
	)
}

func (ch *ColumnHandler) DeleteColumn(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, ok := parseColumnPathParams(ctx)
	if !ok {
		return
	}

	input := column.DeleteColumnInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		RequesterID: userID,
	}

	err := ch.columnUseCase.DeleteColumn(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleColumnError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Column deleted successfully",
		nil,
	)
}

func (ch *ColumnHandler) UpdateColumnPosition(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, ok := parseColumnPathParams(ctx)
	if !ok {
		return
	}

	var req request.UpdateColumnPosition
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := column.UpdateColumnPositionInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Position:    req.Position,
		RequesterID: userID,
	}

	out, err := ch.columnUseCase.UpdateColumnPosition(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleColumnError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Column position updated successfully",
		response.ColumnDTOToResponse(out.Column),
	)
}
