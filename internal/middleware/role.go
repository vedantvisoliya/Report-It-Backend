package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequiredAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRole(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unathorized",
				"ok":    false,
			})
			return
		}

		if !strings.EqualFold(role, "admin") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "this route cna be accessed by admin only",
				"ok":    false,
			})
			return
		}

		c.Next()
	}
}

func RequiredMaster() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRole(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unathorized",
				"ok":    false,
			})
			return
		}

		if !strings.EqualFold(role, "master") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "this route cna be accessed by admin only",
				"ok":    false,
			})
			return
		}

		c.Next()
	}
}
