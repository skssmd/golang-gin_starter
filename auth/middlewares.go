package auth

import (
	"base/initials"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginRequired(c *gin.Context) {
	// Try to get the access token from the cookie
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		// If no access token in cookie, check the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication token missing"})
			c.Abort()
			return
		}

		// Extract the token from the 'Bearer <token>' format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}
		accessToken = authHeader[7:] // Extract the token part
	}

	// Parse and validate the access token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid access token",
		})
		c.Abort()
		return
	}

	// Extract user ID from the access token claims
	var userID string
	switch v := claims["sub"].(type) {
	case string:
		// If the user ID is stored as a string, directly assign it
		userID = v
	case float64:
		// If the user ID is stored as a number (float64), convert it to string
		userID = fmt.Sprintf("%.0f", v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user ID type in token",
		})
		c.Abort()
		return
	}

	// Fetch user from the database
	var user User
	if err := initials.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not found",
		})
		c.Abort()
		return
	}

	// Attach the user to the context for further use in handlers
	c.Set("user", &user)

	// Proceed with the request
	c.Next()
}

func RequireVerification(c *gin.Context) {

	if !IsVerified(c, GetUser(c)) {

		c.JSON(http.StatusForbidden, gin.H{"error": "user is not verified"})
		c.Abort()
		return // `IsVerified` already handles response & abort
	}
	c.Next() // Proceed if user is verified
}

func RequireAdmin(c *gin.Context) {
	if !IsAdmin(c, GetUser(c)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		c.Abort()
		return
	}
	c.Next()
}

func RequireSuperuser(c *gin.Context) {
	// Ensure the user is authenticated

	if !IsSuperuser(c, GetUser(c)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "superuser access required"})
		c.Abort()

		return
	}

	c.Next() // Proceed with request
}
