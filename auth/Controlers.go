package auth

import (
	"base/initials"

	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	var body struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	// Bind the incoming request body to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var errors []string
	var existingUser User

	// Check for existing email
	if body.Email != "" {
		if err := initials.DB.Where("email = ?", body.Email).First(&existingUser).Error; err == nil {
			errors = append(errors, "email already exists")
		}
	}

	// Check for existing username
	if body.Username != "" {
		if err := initials.DB.Where("username = ?", body.Username).First(&existingUser).Error; err == nil {
			errors = append(errors, "username already exists")
		}
	}

	// If any errors exist, return them together
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to hash password"})
		return
	}

	// Create the user
	user := User{
		Email:     body.Email,
		Username:  body.Username,
		Password:  string(hash),
		FirstName: body.FirstName,
		LastName:  body.LastName,
	}

	result := initials.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add user"})
		return
	}

	// Generate JWT tokens using generateJWT function
	accessToken, err := generateJWT(user.ID, time.Hour*24) // 24-hour access token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	refreshToken, err := generateJWT(user.ID, 30*24*time.Hour) // 30-day refresh token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate refresh token"})
		return
	}

	// Set the access token and refresh token in secure HTTP-only cookies
	c.SetCookie("access_token", accessToken, 3600*24, "/", "", true, true)      // 24-hour access token
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", true, true) // 30-day refresh token

	// Send the access token and refresh token in headers for non-browser apps or APIs
	c.Header("Authorization", "Bearer "+accessToken) // Send access token in Authorization header
	c.Header("Refresh-Token", refreshToken)          // Send refresh token in a separate header

	// Generate verification token
	verificationToken, err := generateJWT(user.ID, 24*time.Hour) // 24-hour verification token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate verification token"})
		return
	}

	// Check if email verification is enabled
	if os.Getenv("VERIFICATION") == "true" {
		// Generate verification URL
		verificationURL := fmt.Sprintf("%s/verify-email?token=%s", os.Getenv("SERVER_URL"), verificationToken)

		// Send the verification URL via email
		if err := sendVerificationEmail(user.Email, verificationURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send verification email"})
			return
		}

		// Return success message (no token exposure in response)
		c.JSON(http.StatusOK, gin.H{"message": "User created successfully. Please verify your email."})
	} else {
		// Return verification token in response for internal use
		c.JSON(http.StatusOK, gin.H{
			"message":            "User created successfully, please verify your email.",
			"verification_token": verificationToken,
		})
	}
}
func Login(c *gin.Context) {
	var body struct {
		EmailorUname string `json:"user"`
		Password     string `json:"password"`
	}

	// Bind the incoming request body to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read body",
		})
		return
	}

	var user User

	// Query the user by email or username
	result := initials.DB.Where("email = ? OR username = ?", body.EmailorUname, body.EmailorUname).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user not found",
		})
		return
	}

	// Compare the hashed password with the provided password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// Generate JWT access token using the generateJWT function
	accessToken, err := generateJWT(user.ID, time.Hour*24) // 24-hour access token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate access token",
		})
		return
	}

	// Generate JWT refresh token using the generateJWT function
	refreshToken, err := generateJWT(user.ID, 30*24*time.Hour) // 30-day refresh token
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate refresh token",
		})
		return
	}

	// Set the access token in an HTTP-only cookie (secure, if using HTTPS)
	c.SetCookie("access_token", accessToken, 3600*24, "/", "", true, true) // 24-hour access token

	// Set the refresh token in an HTTP-only cookie (secure, if using HTTPS)
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", true, true) // 30-day refresh token

	// Send the access token and refresh token in headers for non-browser apps or APIs
	c.Header("Authorization", "Bearer "+accessToken) // Send access token in Authorization header
	c.Header("Refresh-Token", refreshToken)          // Send refresh token in a separate header

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message":       "login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
func UserUpdate(c *gin.Context) {
	// Extract user from context (after being authenticated)
	currentUser := GetUser(c)

	if c.Request.Method == "GET" {
		// Send the user information as a response
		c.JSON(http.StatusOK, gin.H{
			"email":      currentUser.Email,
			"username":   currentUser.Username,
			"first_name": currentUser.FirstName,
			"last_name":  currentUser.LastName,
		})
		return
	}

	if c.Request.Method == "POST" {
		// Declare body structure for the update
		var body struct {
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}

		// Bind the incoming request body to the struct
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}

		// Check if the username already exists
		var existingUser User
		if body.Username != "" && body.Username != currentUser.Username {
			if err := initials.DB.Where("username = ?", body.Username).First(&existingUser).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
				return
			}
		}

		// Update only the fields that have been provided
		if body.Username != "" {
			currentUser.Username = body.Username
		}
		if body.FirstName != "" {
			currentUser.FirstName = body.FirstName
		}
		if body.LastName != "" {
			currentUser.LastName = body.LastName
		}

		// Save updated user to the database
		if err := initials.DB.Save(currentUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		// Send success response
		c.JSON(http.StatusOK, gin.H{
			"message": "User information updated successfully",
			"user": gin.H{
				"username":   currentUser.Username,
				"first_name": currentUser.FirstName,
				"last_name":  currentUser.LastName,
			},
		})
		return
	}

	// Handle unsupported HTTP methods
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
}
func DeleteUser(c *gin.Context) {
	// Extract user from context (after being authenticated)
	currentUser := GetUser(c)

	// Declare body structure to receive password input
	var body struct {
		Password string `json:"password"`
	}

	// Bind the incoming request body to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Compare the provided password with the stored hashed password
	err := bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	// Delete the user from the database
	if err := initials.DB.Delete(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	// Optionally, you can clear any session or authentication-related data here
	// For example, clearing cookies or revoking tokens

	// Send success response
	c.JSON(http.StatusOK, gin.H{
		"message": "User account deleted successfully",
	})
	Logout(c)
}

func RefreshToken(c *gin.Context) {
	// First, attempt to get the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		// If no refresh token in the cookie, check the Authorization header
		refreshToken = c.GetHeader("Refresh-Token")
		if refreshToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token not found in cookie or header",
			})
			return
		}
	}

	// Parse and validate the refresh token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	// Extract user ID from the refresh token claims
	var userID string
	switch v := claims["sub"].(type) {
	case string:
		userID = v
	case float64:
		// Convert float64 to string if the user ID is stored as a number
		userID = fmt.Sprintf("%.0f", v)
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user ID type in token",
		})
		return
	}

	// Generate new access token (expires in 1 hour)
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 1).Unix(), // New access token expires in 1 hour
	})

	// Sign the new access token
	newAccessTokenString, err := newAccessToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not generate new access token",
		})
		return
	}

	// Set the new access token in the cookie (secure, if using HTTPS)
	c.SetCookie("access_token", newAccessTokenString, 3600*1, "/", "", true, true)

	// Send the new access token in the Authorization header for non-browser apps
	c.Header("Authorization", "Bearer "+newAccessTokenString)

	// Respond with the new access token
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessTokenString,
	})
}

func Validate(c *gin.Context) {
	user := GetUser(c)
	if user == nil {
		return // Unauthorized response is handled inside GetUser
	}

	c.JSON(http.StatusOK, gin.H{

		"user": user,
		// gin.H{
		// 	"id":       user.ID,
		// 	"email":    user.Email,
		// 	"username": user.Username,

		// },
	})
}

func Logout(c *gin.Context) {
	// Remove the refresh token by setting the cookie expiration time to a past date
	c.SetCookie("access_token", "", -1, "/", "", true, true)  // Setting expiration to -1 will delete the cookie
	c.SetCookie("refresh_token", "", -1, "/", "", true, true) // Setting expiration to -1 will delete the cookie
	// Respond with a success message indicating the user has been logged out
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}
func ChangePassword(c *gin.Context) {
	user := GetUser(c) // Get authenticated user

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	// Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Check if current password is correct
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.CurrentPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect current password"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash new password"})
		return
	}

	// Update password in the database
	user.Password = string(hashedPassword)
	initials.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

func ForgotPassword(c *gin.Context) {
	var body struct {
		EmailOrUsername string `json:"user"`
	}

	// Bind request
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Find user
	var user User
	if err := initials.DB.Where("email = ? OR username = ?", body.EmailOrUsername, body.EmailOrUsername).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	// Generate reset token (valid for 15 minutes)
	resetToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})

	// Sign token
	resetTokenString, err := resetToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate reset token"})
		return
	}

	// Form a reset URL with the token
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("SERVER_URL"), resetTokenString)

	// Check if email verification is enabled
	if os.Getenv("VERIFICATION") == "true" {
		// Send the reset URL via email (using Gomail)
		if err := sendPasswordResetEmail(user.Email, resetURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send reset email"})
			return
		}

		// Send the URL as part of the response
		c.JSON(http.StatusOK, gin.H{
			"message":   "password reset token generated and email sent",
			"reset_url": resetURL,
		})
	} else {
		// If VERIFICATION is false, return only the reset token in the response
		c.JSON(http.StatusOK, gin.H{
			"message": "password reset token generated",
			"token":   resetTokenString,
		})
	}
}

func UpdatePassword(c *gin.Context) {
	var body struct {
		ResetToken  string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	// Bind request
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	// Parse and validate reset token
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(body.ResetToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired reset token"})
		return
	}

	// Extract user ID from token
	userID, ok := claims["sub"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	// Find user
	var user User
	if err := initials.DB.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash new password"})
		return
	}

	// Update password
	user.Password = string(hashedPassword)
	initials.DB.Save(&user)

	// Generate new JWT tokens for login
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	// Sign tokens
	accessTokenString, _ := accessToken.SignedString([]byte(os.Getenv("SECRET")))
	refreshTokenString, _ := refreshToken.SignedString([]byte(os.Getenv("SECRET")))

	// Set refresh token in cookie
	c.SetCookie("Auth", refreshTokenString, 3600*24*30, "/", "", true, true)

	// Response
	c.JSON(http.StatusOK, gin.H{
		"message":       "password updated successfully",
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})
}
func VerifyEmail(c *gin.Context) {
	var body struct {
		VerificationToken string `json:"verification_token"`
	}

	// Bind the incoming request body to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read body",
		})
		return
	}

	// Parse the verification token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(body.VerificationToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid or expired verification token",
		})
		return
	}

	// Check if the token has expired by comparing the 'exp' claim with the current time
	expiration, ok := claims["exp"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token payload",
		})
		return
	}

	if time.Now().Unix() > int64(expiration) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "verification token has expired",
		})
		return
	}

	// Extract user ID from token
	userID, ok := claims["sub"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token payload",
		})
		return
	}

	// Find the user
	var user User
	if err := initials.DB.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user not found",
		})
		return
	}

	// Update the user to set IsVerified to true
	user.IsVerified = true
	if err := initials.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update user verification status",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"message": "email verified successfully",
	})
}
func ResendVerificationToken(c *gin.Context) {
	// Fetch the authenticated user
	user := GetUser(c) // Get the user from the context (Assumes user is authenticated)
	if user == nil {
		return // If user is nil, GetUser already returned an unauthorized response
	}

	// Check if the user is already verified
	if user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already verified"})
		return
	}

	// Generate a new verification token (valid for 24 hours)
	verificationToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,                               // User ID is the subject
		"exp": time.Now().Add(24 * time.Hour).Unix(), // Token expiration (24 hours)
	})

	// Sign the verification token
	tokenString, err := verificationToken.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate verification token"})
		return
	}

	// Check if email verification is enabled
	if os.Getenv("VERIFICATION") == "true" {
		// Send the token via email
		if err := sendVerificationEmail(user.Email, tokenString); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
			return
		}

		// Success response
		c.JSON(http.StatusOK, gin.H{"message": "Verification token sent to email"})
	} else {
		// Return the token as JSON if email verification is not enabled
		c.JSON(http.StatusOK, gin.H{
			"message":            "Verification token generated",
			"verification_token": tokenString,
		})
	}
}

// GetVerifiedAndUnverifiedUsers retrieves verified and unverified users separately
func GetVerifiedAndUnverifiedUsers(c *gin.Context) {

	var users []User
	var verifiedUsers []User
	var unverifiedUsers []User

	// Fetch all users from the database
	if err := initials.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Iterate through users and use IsVerified method
	for _, user := range users {
		if IsVerified(c, &user) {
			verifiedUsers = append(verifiedUsers, user)
		} else {
			unverifiedUsers = append(unverifiedUsers, user)
		}
	}

	// Return response in JSON
	c.JSON(http.StatusOK, gin.H{
		"verified_users":   verifiedUsers,
		"unverified_users": unverifiedUsers,
	})
}
