package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		_ = start
		_ = path
		_ = method

		// TODO: plug in real logger here (status, latency, etc.)
	}
}
