package response

import (
	"fmt"
	"net/http"

	"collabotask/internal/adapter/http/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func HandleValidationError(c *gin.Context, err error) {
	details := []string{}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			details = append(details, fmt.Sprintf("%s: %s", e.Field(), e.Tag()))
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
