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
