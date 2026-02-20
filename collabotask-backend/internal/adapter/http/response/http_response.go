package response

import (
	"collabotask/internal/adapter/http/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func GenerateSuccessResponse(c *gin.Context, message string, data interface{}, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	c.JSON(code, APIResponse{
		StatusCode: code,
		Message:    message,
		Data:       data,
		Success:    true,
	})
}

func GenerateErrorResponse(c *gin.Context, err error) {
	if e, ok := err.(*errors.AppError); ok {
		c.JSON(e.StatusCode, APIResponse{
			Success:    false,
			StatusCode: e.StatusCode,
			Message:    e.Message,
			Error: &APIError{
				Code:    e.Code,
				Message: e.Message,
			},
		})

		return
	}

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success:    false,
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal server error",
		Error: &APIError{
			Code:    errors.ErrCodeInternal,
			Message: err.Error(),
		},
	})
}

func GenerateDetailedErrorResponse(c *gin.Context, statusCode int, message string, apiErr *APIError) {
	c.JSON(statusCode, APIResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Error:      apiErr,
	})
}
