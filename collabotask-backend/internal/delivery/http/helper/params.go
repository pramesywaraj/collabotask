package helper

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUUIDParams(ctx *gin.Context, param string) (uuid.UUID, bool) {
	s := ctx.Param(param)
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}
