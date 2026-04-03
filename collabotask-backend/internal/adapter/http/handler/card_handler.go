package handler

import (
	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/helper"
	"collabotask/internal/adapter/http/request"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/domain"
	"collabotask/internal/usecase/card"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CardHandler struct {
	cardUseCase card.CardUseCase
}

func NewCardHandler(cu card.CardUseCase) *CardHandler {
	return &CardHandler{
		cardUseCase: cu,
	}
}

func handleCardError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotInWorkspace),
		errors.Is(err, domain.ErrBoardAccessDenied),
		errors.Is(err, domain.ErrBoardPermissionDenied),
		errors.Is(err, domain.ErrNotWorkspaceAdmin):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusForbidden, apperrors.ErrCodeForbidden, err.Error()))
	case errors.Is(err, domain.ErrBoardNotFound),
		errors.Is(err, domain.ErrColumnNotFound),
		errors.Is(err, domain.ErrCardNotFound),
		errors.Is(err, domain.ErrWorkspaceNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrMemberNotFound),
		errors.Is(err, domain.ErrBoardMemberNotFound),
		errors.Is(err, domain.ErrColumnNotInBoard),
		errors.Is(err, domain.ErrCardNotInColumn):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusNotFound, apperrors.ErrCodeNotFound, err.Error()))
	case errors.Is(err, domain.ErrConstraintViolation),
		errors.Is(err, domain.ErrAtLeastOneProvided),
		errors.Is(err, domain.ErrInvalidAssigneeID):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, err.Error()))
	case errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrBoardAlreadyMember),
		errors.Is(err, domain.ErrInconsistentState):
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusConflict, apperrors.ErrCodeConflict, err.Error()))
	default:
		response.GenerateErrorResponse(ctx, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, err.Error()))
	}
}

func parseCardPathParams(ctx *gin.Context) (boardID, columnID, cardID uuid.UUID, ok bool) {
	boardID, okBoard := helper.ParseUUIDParams(ctx, "board_id")
	columnID, okColumn := helper.ParseUUIDParams(ctx, "column_id")
	cardID, okCard := helper.ParseUUIDParams(ctx, "card_id")
	if !okBoard || !okColumn || !okCard {
		var errMessage string
		switch {
		case !okBoard && !okColumn && !okCard:
			errMessage = "Invalid or missing board, column, and card ids"
		case !okBoard && !okColumn:
			errMessage = "Invalid or missing board and column ids"
		case !okBoard && !okCard:
			errMessage = "Invalid or missing board and card ids"
		case !okColumn && !okCard:
			errMessage = "Invalid or missing column and card ids"
		case !okBoard:
			errMessage = "Invalid or missing board id"
		case !okColumn:
			errMessage = "Invalid or missing column id"
		default:
			errMessage = "Invalid or missing card id"
		}
		response.GenerateErrorResponse(
			ctx,
			apperrors.NewAppError(http.StatusBadRequest, apperrors.ErrCodeValidation, errMessage),
		)
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return boardID, columnID, cardID, true
}

func parseBoardAndColumnPathParams(ctx *gin.Context) (boardID, columnID uuid.UUID, ok bool) {
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

// CreateCard godoc
// @Summary Create a card in a column
// @Tags card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param column_id path string true "Column UUID"
// @Param body body request.CreateCardRequest true "Card payload"
// @Success 201 {object} response.CardCreateSuccessDoc "Created"
// @Failure 400 {object} response.Failure400ValidationDoc "Invalid board/column id or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/columns/{column_id}/cards [post]
func (crh *CardHandler) CreateCard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, ok := parseBoardAndColumnPathParams(ctx)
	if !ok {
		return
	}

	var req request.CreateCardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := card.CreateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       req.Title,
		Description: req.Description,
		AssignedTo:  req.AssignedTo,
		DueDate:     req.DueDate,
		RequesterID: userID,
	}

	out, err := crh.cardUseCase.CreateCard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleCardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Card created successfully",
		response.CardDTOToResponse(out.Card),
		http.StatusCreated,
	)
}

// UpdateCard godoc
// @Summary Update a card
// @Tags card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param column_id path string true "Column UUID"
// @Param card_id path string true "Card UUID"
// @Param body body request.UpdateCardRequest true "Partial update"
// @Success 200 {object} response.CardUpdateSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Invalid ids or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/columns/{column_id}/cards/{card_id} [patch]
func (crh *CardHandler) UpdateCard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, cardID, ok := parseCardPathParams(ctx)
	if !ok {
		return
	}

	var req request.UpdateCardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := card.UpdateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		CardID:      cardID,
		Title:       req.Title,
		RequesterID: userID,
	}
	if req.Description.Present {
		input.DescriptionPresent = true
		input.Description = req.Description.Value
	}
	if req.AssignedTo.Present {
		input.AssignedToPresent = true
		input.AssignedTo = req.AssignedTo.Value
	}
	if req.DueDate.Present {
		input.DueDatePresent = true
		input.DueDate = req.DueDate.Value
	}

	out, err := crh.cardUseCase.UpdateCard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleCardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Card updated successfully",
		response.CardDTOToResponse(out.Card),
	)
}

// DeleteCard godoc
// @Summary Delete a card
// @Tags card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param column_id path string true "Column UUID"
// @Param card_id path string true "Card UUID"
// @Success 200 {object} response.CardDeleteSuccessDoc "OK"
// @Failure 400 {object} response.Failure400BadRequestDoc "Invalid board/column/card id"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/columns/{column_id}/cards/{card_id} [delete]
func (crh *CardHandler) DeleteCard(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, cardID, ok := parseCardPathParams(ctx)
	if !ok {
		return
	}

	input := card.DeleteCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		CardID:      cardID,
		RequesterID: userID,
	}

	err := crh.cardUseCase.DeleteCard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleCardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Card deleted successfully",
		nil,
	)
}

// MoveCardPosition godoc
// @Summary Move a card (column and/or position)
// @Tags card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "Workspace UUID"
// @Param board_id path string true "Board UUID"
// @Param column_id path string true "Source column UUID"
// @Param card_id path string true "Card UUID"
// @Param body body request.MoveCardRequest true "Target column and position"
// @Success 200 {object} response.CardMoveSuccessDoc "OK"
// @Failure 400 {object} response.Failure400ValidationDoc "Invalid ids or validation error"
// @Failure 401 {object} response.Failure401UnauthorizedDoc "Unauthorized"
// @Failure 403 {object} response.Failure403ForbiddenDoc "Forbidden"
// @Failure 404 {object} response.Failure404NotFoundDoc "Not found"
// @Failure 409 {object} response.Failure409ConflictDoc "Conflict"
// @Failure 500 {object} response.Failure500InternalDoc "Internal server error"
// @Router /workspace/{workspace_id}/board/{board_id}/columns/{column_id}/cards/{card_id}/move [post]
func (crh *CardHandler) MoveCardPosition(ctx *gin.Context) {
	userID, ok := helper.GetAndCheckUserID(ctx)
	if !ok {
		return
	}

	boardID, columnID, cardID, ok := parseCardPathParams(ctx)
	if !ok {
		return
	}

	var req request.MoveCardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(ctx, err)
		return
	}

	input := card.MoveCardInput{
		BoardID:      boardID,
		CardID:       cardID,
		FromColumnID: columnID,
		ToColumnID:   req.ToColumnID,
		ToPosition:   req.ToPosition,
		RequesterID:  userID,
	}

	out, err := crh.cardUseCase.MoveCard(ctx.Request.Context(), input)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			response.HandleValidationError(ctx, err)
			return
		}

		handleCardError(ctx, err)
		return
	}

	response.GenerateSuccessResponse(
		ctx,
		"Card position updated successfully",
		response.CardDTOToResponse(out.Card),
	)
}
