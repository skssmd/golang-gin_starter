package auth

import (
	"github.com/gin-gonic/gin"
)

// Routes defines all authentication-related routes.
func Routes(r *gin.Engine) {
	// Parent route /auth
	authGroup := r.Group("/auth")
	{
		// Signup route
		authGroup.POST("/signup", Signup)

		// Login route
		authGroup.POST("/login", Login)

		// Validate route (protected)
		authGroup.GET("/validate", LoginRequired, RequireVerification, Validate)

		// Logout route (removes the refresh token)
		authGroup.POST("/logout", Logout)

		// Refresh token route (to get a new access token using the refresh token)
		authGroup.POST("/refresh-token", RefreshToken)

		authGroup.POST("/change-password", LoginRequired, ChangePassword)
		authGroup.POST("/forgot-password", ForgotPassword)
		authGroup.POST("/reset-password", UpdatePassword)
		authGroup.POST("/verify-email", VerifyEmail)
		authGroup.POST("/resend-verification-token", LoginRequired, ResendVerificationToken)
		authGroup.GET("/google/login", GoogleLogin)
		authGroup.GET("/google/callback", GoogleCallback)
	}
	userroutes := r.Group("/user", LoginRequired) // Apply LoginRequired middleware to the whole group
	{
		userroutes.GET("/update", UserUpdate)    // GET request to retrieve user info
		userroutes.POST("/update", UserUpdate)   // POST request to update user info
		userroutes.DELETE("/delete", DeleteUser) // DELETE request to delete user
	}
	superuserroutes := r.Group("/controlpanel", LoginRequired, RequireSuperuser) // Apply LoginRequired middleware to the whole group
	{
		superuserroutes.GET("/users", GetVerifiedAndUnverifiedUsers) // GET request to retrieve user info
	}

}
