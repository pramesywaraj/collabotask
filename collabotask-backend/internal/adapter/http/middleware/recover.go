package middleware

import (
	"fmt"
	"net/http"

	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/response"
	"collabotask/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recover(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorWithErr("panic recovered", fmt.Errorf("%v", r))
				response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusInternalServerError, apperrors.ErrCodeInternal, "Internal server error"))
				c.Abort()
			}
		}()

		c.Next()
	}
}
