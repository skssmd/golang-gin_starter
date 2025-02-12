package subscription

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"base/auth"     // Import the auth package to access GetUser
	"base/initials" // Import the initials package where Db() is defined

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"gorm.io/gorm"
)

func HandleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	bodyReader := io.LimitReader(c.Request.Body, MaxBodyBytes) // Limit the size of the body for security
	payload, err := io.ReadAll(bodyReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		c.JSON(503, gin.H{"error": "Service Unavailable"})
		return
	}

	// Replace with your actual Stripe webhook secret key
	endpointSecret := "whsec_12345"
	signatureHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Webhook signature verification failed. %v\n", err)
		c.JSON(400, gin.H{"error": "Webhook signature verification failed"})
		return
	}

	// Initialize the database connection
	initials.Db()     // Ensure that the database connection is established
	db := initials.DB // Use the global DB from the initials package

	// Get the current user
	currentUser := auth.GetUser(c)
	if currentUser == nil {
		return // If user is not authenticated, exit early
	}

	// User ID to associate with the subscription
	userID := currentUser.ID

	switch event.Type {
	case "customer.subscription.created":
		var stripeSubscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSubscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			c.JSON(400, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// Handle subscription creation
		handleSubscriptionCreated(db, userID, stripeSubscription)

	case "customer.subscription.updated":
		var stripeSubscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSubscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			c.JSON(400, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// Handle subscription update
		handleSubscriptionUpdated(db, stripeSubscription)

	case "customer.subscription.deleted":
		var stripeSubscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSubscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			c.JSON(400, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// Handle subscription deletion
		handleSubscriptionDeleted(db, stripeSubscription)

	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	c.JSON(200, gin.H{"status": "success"})
}

func handleSubscriptionCreated(db *gorm.DB, userID uint, stripeSubscription stripe.Subscription) {
	// Convert stripe.SubscriptionStatus to string
	status := string(stripeSubscription.Status) // Convert status to string

	// Convert Unix timestamp to time.Time (if TrialEnd is non-zero)
	var trialEndDate time.Time
	if stripeSubscription.TrialEnd != 0 {
		trialEndDate = time.Unix(stripeSubscription.TrialEnd, 0) // Convert the Unix timestamp to time.Time
	}

	// Create a new subscription in the database
	newSubscription := Subscription{
		UserID:               userID, // Use the userID from GetUser
		StripeSubscriptionID: stripeSubscription.ID,
		Status:               status,                                    // Use the converted string
		PlanID:               stripeSubscription.Items.Data[0].Price.ID, // Accessing the first subscription item
		StartDate:            time.Now(),                                // Adjust based on the Stripe data
		EndDate:              time.Now().Add(30 * 24 * time.Hour),       // Assuming a monthly plan
		TrialEndDate:         trialEndDate,                              // Use the converted time.Time
		PaymentMethodID:      stripeSubscription.DefaultPaymentMethod.ID,
	}
	db.Create(&newSubscription) // Assuming `db` is your GORM database connection
}

func handleSubscriptionUpdated(db *gorm.DB, stripeSubscription stripe.Subscription) {
	// Convert status to string (if necessary)
	status := string(stripeSubscription.Status)

	// Convert Unix timestamp to time.Time (if TrialEnd is non-zero)
	var trialEndDate time.Time
	if stripeSubscription.TrialEnd != 0 {
		trialEndDate = time.Unix(stripeSubscription.TrialEnd, 0)
	}

	// Update the subscription in the database
	db.Model(&Subscription{}).Where("stripe_subscription_id = ?", stripeSubscription.ID).
		Updates(map[string]interface{}{
			"status":         status,       // Updated status
			"trial_end_date": trialEndDate, // Updated trial end date
		})
}

func handleSubscriptionDeleted(db *gorm.DB, stripeSubscription stripe.Subscription) {
	// Delete the subscription from the database
	db.Delete(&Subscription{}, "stripe_subscription_id = ?", stripeSubscription.ID)
}
