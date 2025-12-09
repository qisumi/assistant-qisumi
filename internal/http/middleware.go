package http

import (
	"net/http"
	"strings"

	"assistant-qisumi/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		userID, err := jwtMgr.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get("userID"); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}
