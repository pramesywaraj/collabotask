package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"collabotask/internal/config"

	"github.com/gin-gonic/gin"
)

func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if len(cfg.AllowedOrigins) > 0 {
			if slices.Contains(cfg.AllowedOrigins, "*") {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && slices.Contains(cfg.AllowedOrigins, origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(cfg.AllowedOrigins) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.AllowedOrigins[0])
			}
		}

		if len(cfg.AllowedMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		}
		if len(cfg.AllowedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		}
		if cfg.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if cfg.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
