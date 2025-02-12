package auth

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) *User {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return nil
	}

	// Convert interface{} to *User
	currentUser, ok := user.(*User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user data"})
		c.Abort()
		return nil
	}

	return currentUser
}

// Check if the user is verified based on the environment variable
func IsVerified(c *gin.Context, q *User) bool {
	user := q
	if user == nil {
		return false
	}

	verificationEnabled := os.Getenv("VERIFICATION") == "true"
	if verificationEnabled && !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "user is not verified"})
		c.Abort()
		return false
	}

	return user.IsVerified
}

// Check if the user is an admin, and if verification is required, check it
func IsAdmin(c *gin.Context, q *User) bool {
	user := q
	if user == nil {
		return false
	}

	verificationEnabled := os.Getenv("VERIFICATION") == "true"
	if verificationEnabled && !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access requires verified user"})
		c.Abort()
		return false
	}

	return user.IsAdmin
}

// Check if the user is a superuser, and if verification is required, check it
func IsSuperuser(c *gin.Context, q *User) bool {
	fmt.Println(c)
	user := q
	if user == nil {
		return false
	}

	verificationEnabled := os.Getenv("VERIFICATION") == "true"
	if verificationEnabled && !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "superuser access requires verified user"})
		c.Abort()
		return false
	}

	return user.IsSuperuser
}
