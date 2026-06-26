package helper

import (
	apperrors "collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/response"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func HandleUseCaseError(ctx *gin.Context, err error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		response.HandleValidationError(ctx, err)
		return
	}

	response.GenerateErrorResponse(ctx, apperrors.MapDomainError(err))
}
