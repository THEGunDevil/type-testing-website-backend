package middleware

import (
	"net/http"
	"strings"
	"typetype/internal/utils/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Try to get the access token from the httpOnly cookie first
		cookieToken, err := c.Cookie("access_token")
		if err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// 2. Fallback to Authorization header if cookie is missing
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: missing token",
			})
			c.Abort()
			return
		}

		// Verify token
		claims, err := jwtManager.VerifyAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(), // "token has expired" or "invalid token"
			})
			c.Abort()
			return
		}

		// Inject data into context
		c.Set("userID", claims.ID)
		c.Set("userRole", claims.Role)
		c.Set("tokenVersion", claims.TokenVersion)
		c.Next()
	}
}