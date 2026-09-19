package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"live-polling-tool/backend/utils"
	"net/http"
	"strings"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token := strings.TrimPrefix(h, "Bearer ")
		if token == h {
			token, _ = c.Cookie("auth_token")
		}
		if token == "" {
			utils.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			c.Abort()
			return
		}
		t, e := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if e != nil || !t.Valid {
			utils.Error(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}
		sub, e := t.Claims.GetSubject()
		if e != nil || sub == "" {
			utils.Error(c, 401, "UNAUTHORIZED", "Invalid token")
			c.Abort()
			return
		}
		c.Set("userID", sub)
		c.Next()
	}
}
