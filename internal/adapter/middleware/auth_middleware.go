package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtService *service.JWTService
}

func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a middleware that validates JWT tokens
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		// 1. Try getting from cookie
		cookieToken, err := c.Cookie("access_token")
		if err == nil && cookieToken != "" {
			tokenString = cookieToken
		}

		// 2. Fallback to Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(dto.ErrCodeUnauthorized, "Authentication required")))
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			errCode := dto.ErrCodeInvalidToken
			if errors.Is(err, jwt.ErrTokenExpired) {
				errCode = dto.ErrCodeExpiredToken
			}

			c.JSON(http.StatusUnauthorized, dto.NewAPIResponse[any](nil,
				dto.NewErrorResponse(errCode)))
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
