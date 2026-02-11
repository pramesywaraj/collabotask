package middleware

import "github.com/gin-gonic/gin"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: extract JWT from Authorization header, validate, and set userID in context
		c.Next()
	}
}
