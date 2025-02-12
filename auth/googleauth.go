package auth

import (
	"base/initials"
	"context"
	"log"
	"net/http"

	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// OAuth2 config for Google
var googleOAuthConfig = &oauth2.Config{
	ClientID:     "Goole_Client_ID",
	ClientSecret: "Google_client_secret",
	RedirectURL:  "http://127.0.0.1:3000/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// GoogleLogin - Redirects to Google login
func GoogleLogin(c *gin.Context) {
	url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	c.Redirect(http.StatusTemporaryRedirect, url)

}

// GoogleCallback - Handles Google OAuth callback
func GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not found"})
		return
	}

	// Exchange code for token
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Token exchange error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Create an OAuth2 client
	client := googleOAuthConfig.Client(context.Background(), token)

	// Fetch user info
	oauth2Service, err := oauth2v2.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Println("OAuth2 service creation failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OAuth2 service"})
		return
	}

	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		log.Println("Failed to retrieve user info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// Extract first and last name
	names := strings.SplitN(userinfo.Name, " ", 2)
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = names[1]
	}

	// Check if user exists in database
	var user User
	if err := initials.DB.Where("email = ?", userinfo.Email).First(&user).Error; err != nil {
		// Create new user
		user = User{
			Email:      userinfo.Email,
			Username:   userinfo.Email,
			FirstName:  firstName,
			LastName:   lastName,
			IsVerified: true, // Google users are automatically verified
		}
		if err := initials.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// Update missing first/last name if necessary
		updates := map[string]interface{}{}
		if user.FirstName == "" {
			updates["first_name"] = firstName
		}
		if user.LastName == "" {
			updates["last_name"] = lastName
		}
		if len(updates) > 0 {
			initials.DB.Model(&user).Updates(updates)
		}
	}

	// Generate JWT Access Token (short-lived)
	accessToken, err := generateJWT(user.ID, time.Hour*24) // 24-hour validity
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// Generate JWT Refresh Token (long-lived)
	refreshToken, err := generateJWT(user.ID, time.Hour*24*30) // 30-day validity
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Store refresh token in an HTTP-only cookie (secure & HTTP-only)
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", true, true)

	// Store access token in an HTTP-only cookie (secure & HTTP-only)
	c.SetCookie("access_token", accessToken, 3600*24, "/", "", true, true)
	c.Header("Authorization", "Bearer "+accessToken) // Send access token in Authorization header
	c.Header("Refresh-Token", refreshToken)          // Send refresh token in a separate header
	// Respond with user details and tokens in response body and headers
	c.JSON(http.StatusOK, gin.H{
		"message": "Google login successful",
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"is_verified": user.IsVerified,
		},
		// Include tokens in the response body for client-side use
		"access_token":  accessToken,
		"refresh_token": refreshToken, // Send refresh token (optional)
	})

	// Optionally, you can also set the tokens in headers, but typically it's better to use cookies for security reasons.

}
