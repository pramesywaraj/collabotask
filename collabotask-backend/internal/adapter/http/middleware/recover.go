package middleware

import (
	"fmt"
	"net/http"

	"collabotask/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recover(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorWithErr("panic recovered", fmt.Errorf("%v", r))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "internal server error",
				})
			}
		}()

		c.Next()
	}
}
