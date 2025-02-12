package subscription

import (
	"base/auth"

	"github.com/gin-gonic/gin"
)

// InitializeSubscriptionRoutes defines all subscription-related routes.
func Routes(r *gin.Engine) {
	// Parent route /subscription
	subscriptionGroup := r.Group("/subscription")
	{
		// Webhook route for Stripe subscription events (without authentication)
		subscriptionGroup.POST("/webhook", HandleWebhook)

		// Route for retrieving available plans (public access)
		subscriptionGroup.GET("/plans", GetPlans)

		// Add authentication middleware for other subscription-related routes
		subscriptionGroup.Use(auth.LoginRequired)
		{
			// Route for creating a subscription (authenticated users only)
			subscriptionGroup.POST("/create", CreateSubscription)

			// Route for handling payment success (authenticated users only)
			subscriptionGroup.GET("/payment-success", PaymentSuccess)
		}
	}
}
