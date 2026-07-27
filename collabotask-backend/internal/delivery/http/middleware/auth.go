package middleware

import (
	"net/http"

	"collabotask/internal/config"
	apperrors "collabotask/internal/delivery/http/errors"
	"collabotask/internal/delivery/http/response"
	infraauth "collabotask/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ContextUserIDKey = "userID"

func Auth(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(cfg.CookieName)
		if err != nil {
			response.GenerateErrorResponse(c, apperrors.NewAppError(http.StatusUnauthorized, apperrors.ErrCodeUnauthorized, "Authentication required"))
			c.Abort()
			return
		}

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
