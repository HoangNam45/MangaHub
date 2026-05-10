package middleware

import (
	"net/http"
	"strings"

	"mangahub/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for valid JWT token
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        // Extract token 
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]
        claims, err := utils.VerifyToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }


        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}