package helper

import (
	"collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/middleware"
	"collabotask/internal/delivery/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAndCheckUserID(ctx *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.GenerateErrorResponse(ctx, errors.NewAppError(http.StatusUnauthorized, errors.ErrCodeUnauthorized, "Unauthorized"))
		return uuid.Nil, false
	}

	return userID, ok
}
