package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := uuid.NewString()

		ctx := context.WithValue(
			c.Request.Context(),
			RequestIDKey,
			rid,
		)

		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set("X-Request-ID", rid)

		c.Next()
	}
}
