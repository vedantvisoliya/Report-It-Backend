package middleware

import (
	"net/http"
	"reportit-api/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ctxUserIDKey = "auth.userID"
	ctxUserRoleKey   = "auth.role"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization token",
				"ok": false,
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalud authorization format",
				"ok": false,
			})
			return
		}

		scheme := strings.TrimSpace(parts[0])
		tokenString := strings.TrimSpace(parts[1])

		if !strings.EqualFold(scheme, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization scheme must be a Bearer",
				"ok": false,
			})
			return
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing access token",
				"ok": false,
			})
			return
		}

		claims, err := auth.ParseToken(jwtSecret, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"ok": false,
			})
			return
		}
		c.Set(ctxUserIDKey, claims.Subject)
		c.Set(ctxUserRoleKey, claims.Role)

		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	res, ok := c.Get(ctxUserIDKey)
	if !ok {
		return "", false
	}

	userID, ok := res.(string)
	return userID, ok
}

func GetUserRole(c *gin.Context) (string, bool) {
	res, ok := c.Get(ctxUserRoleKey)
	if !ok {
		return "", false
	}

	userRole, ok := res.(string)
	return userRole, ok
}