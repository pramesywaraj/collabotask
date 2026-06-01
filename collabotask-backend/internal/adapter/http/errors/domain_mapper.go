package errors

import (
	"collabotask/internal/domain"
	"errors"
	"net/http"
)

func MapDomainError(err error) *AppError {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return NewAppError(http.StatusUnauthorized, ErrCodeUnauthorized, err.Error())
	case errors.Is(err, domain.ErrUserNotInWorkspace),
		errors.Is(err, domain.ErrBoardAccessDenied),
		errors.Is(err, domain.ErrBoardPermissionDenied),
		errors.Is(err, domain.ErrBoardOwnerCannotLeave),
		errors.Is(err, domain.ErrNotWorkspaceAdmin):
		return NewAppError(http.StatusForbidden, ErrCodeForbidden, err.Error())
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrMemberNotFound),
		errors.Is(err, domain.ErrWorkspaceNotFound),
		errors.Is(err, domain.ErrBoardNotFound),
		errors.Is(err, domain.ErrBoardMemberNotFound),
		errors.Is(err, domain.ErrColumnNotFound),
		errors.Is(err, domain.ErrColumnNotInBoard),
		errors.Is(err, domain.ErrCardNotFound),
		errors.Is(err, domain.ErrCardNotInColumn):
		return NewAppError(http.StatusNotFound, ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrConstraintViolation),
		errors.Is(err, domain.ErrAtLeastOneProvided),
		errors.Is(err, domain.ErrCannotRemoveYourself),
		errors.Is(err, domain.ErrBoardNoMembersToInvite),
		errors.Is(err, domain.ErrInvalidAssigneeID):
		return NewAppError(http.StatusBadRequest, ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrBoardAlreadyMember),
		errors.Is(err, domain.ErrBoardCannotJoin),
		errors.Is(err, domain.ErrInconsistentState):
		return NewAppError(http.StatusConflict, ErrCodeConflict, err.Error())
	default:
		return NewAppError(http.StatusInternalServerError, ErrCodeInternal, err.Error())
	}
}
