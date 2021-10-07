package middleware

import (
	"errors"
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if adminInfo, ok := session.Get(public.AdminSessionInfoKey).(string); !ok || adminInfo == "" {
			ResponseError(c, InternalErrorCode, errors.New("User not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
