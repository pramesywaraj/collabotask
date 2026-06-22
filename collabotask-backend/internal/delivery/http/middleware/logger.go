package middleware

import (
	"time"

	"collabotask/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		log.WithFields(map[string]interface{}{
			"method":    method,
			"path":      path,
			"status":    status,
			"latency":   latency.Milliseconds(),
			"client_ip": clientIP,
		}).Info("request")
	}
}
