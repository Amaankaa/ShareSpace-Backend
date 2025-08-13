package infrastructure

import (
    "net/http"
    "strings"
    domain "github.com/Amaankaa/Blog-Starter-Project/Domain/user"


    "github.com/gin-gonic/gin"
)


type AuthMiddleware struct {
    jwtService domain.IJWTService
}


func NewAuthMiddleware(jwtService domain.IJWTService) *AuthMiddleware {
    return &AuthMiddleware{
        jwtService: jwtService,
    }
}


func (am *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
	
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
			c.Abort()
			return
		}

        tokenString := strings.TrimPrefix(header, "Bearer ")
        claims, err := am.jwtService.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            c.Abort()
            return
        }


        c.Set("user_id", claims["_id"])
        c.Set("username", claims["username"])
        c.Set("role", claims["role"])
        c.Next()
    }
}


func (am *AuthMiddleware) AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
