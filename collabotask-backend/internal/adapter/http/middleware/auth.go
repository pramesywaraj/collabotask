package middleware

import (
	"net/http"
	"strings"

	apperrors "collabotask/internal/adapter/http/errors"
	"collabotask/internal/adapter/http/response"
	"collabotask/internal/config"
	infraauth "collabotask/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ContextUserIDKey = "userID"

func Auth(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, "Authorization header required"))
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, "Invalid authorization formata"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, prefix)
		claims, err := infraauth.ValidateToken(cfg, tokenString)
		if err != nil {
			response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, "Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := v.(uuid.UUID)
	return userID, ok
}
