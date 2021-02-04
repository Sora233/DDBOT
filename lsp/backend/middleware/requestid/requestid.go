package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const Key = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.GetString(Key)) == 0 {
			c.Set(Key, uuid.New().String())
		}
	}
}

func GetRequestID(ctx *gin.Context) string {
	return ctx.GetString(Key)
}
