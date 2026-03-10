package middleware

import (
	"strings"

	jwtPkg "goaltrack/pkg/jwt"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

const ContextUserID = "user_id"

func JWTAuth(jm *jwtPkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			return
		}

		claims, err := jm.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			return
		}

		if claims.Type != "access" {
			response.Unauthorized(c, "invalid token type")
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(ContextUserID)
	uid, _ := id.(uint)
	return uid
}
