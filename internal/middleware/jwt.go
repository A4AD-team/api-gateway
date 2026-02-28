package middleware

import (
	"crypto/sha256"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	secret string
}

func NewJWTMiddleware(secret string) *JWTMiddleware {
	return &JWTMiddleware{secret: secret}
}

func (m *JWTMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		userID := extractUserID(claims)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user id not found in token"})
			return
		}

		c.Set("user_id", userID)
		c.Set("x_user_id", strconv.FormatInt(userID, 10))
		c.Set("x_username", extractUsername(claims))

		c.Next()
	}
}

func extractUserID(claims jwt.MapClaims) int64 {
	if sub, ok := claims["sub"].(string); ok {
		if id, err := strconv.ParseInt(sub, 10, 64); err == nil {
			return id
		}
		hash := sha256.Sum256([]byte(sub))
		id := int64(hash[0])<<24 | int64(hash[1])<<16 | int64(hash[2])<<8 | int64(hash[3])
		if id < 0 {
			id = -id
		}
		if id == 0 {
			id = int64(hash[4])<<24 | int64(hash[5])<<16 | int64(hash[6])<<8 | int64(hash[7])
			if id < 0 {
				id = -id
			}
		}
		return id
	}
	if sub, ok := claims["sub"].(float64); ok {
		return int64(sub)
	}
	if userID, ok := claims["user_id"].(float64); ok {
		return int64(userID)
	}
	if userID, ok := claims["user_id"].(string); ok {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			return id
		}
	}
	if id, ok := claims["id"].(float64); ok {
		return int64(id)
	}
	return 0
}

func extractUsername(claims jwt.MapClaims) string {
	if username, ok := claims["username"].(string); ok {
		return username
	}
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		return preferredUsername
	}
	if name, ok := claims["name"].(string); ok {
		return name
	}
	if email, ok := claims["email"].(string); ok {
		return email
	}
	return ""
}
