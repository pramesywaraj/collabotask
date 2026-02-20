package response

import (
	"fmt"
	"net/http"

	"collabotask/internal/adapter/http/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var tagToMessage = map[string]string{
	"required": "is required",
	"email":    "must be a valid email address",
	"min":      "must be at least %s characters",
	"max":      "must be at most %s characters",
	"gte":      "must be at least %s",
	"lte":      "must be at most %s",
	"oneof":    "must be one of: %s",
	"alpha":    "must contain only letters",
	"alphanum": "must contain only letters and numbers",
	"url":      "must be a valid URL",
	"uuid":     "must be a valid UUID",
}

func getValidationMessage(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	msg, ok := tagToMessage[tag]
	if !ok {
		return fmt.Sprintf("%s: invalid", field)
	}

	if param != "" {
		return fmt.Sprintf("%s %s", field, fmt.Sprintf(msg, param))
	}
	return fmt.Sprintf("%s %s", field, msg)
}

func HandleValidationError(c *gin.Context, err error) {
	details := []string{}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			details = append(details, getValidationMessage(e))
		}
	} else {
		details = []string{err.Error()}
	}

	GenerateDetailedErrorResponse(c, http.StatusBadRequest, "Validation failed", &APIError{
		Code:    errors.ErrCodeValidation,
		Message: "Request validation failed",
		Details: details,
	})
	return
}
