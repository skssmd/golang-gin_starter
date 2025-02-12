package subscription

import (
	"base/auth" // Import the User model from the auth package
	"base/initials"
	"os"
	"time"

	"gorm.io/gorm"

	"log"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/price"
)

func GenerateStripePaymentMethods() error {
	// Initialize Stripe client with your secret key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// Step 1: Retrieve all users
	var users []auth.User
	err := initials.DB.Find(&users).Error
	if err != nil {
		return err // Return error if there is any issue
	}

	// Step 2: Loop through each user and create a PaymentMethod if needed
	for _, user := range users {
		var existingPaymentMethod PaymentMethod
		err = initials.DB.Where("user_id = ?", user.ID).First(&existingPaymentMethod).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err // Return error if there is an issue while checking PaymentMethod
		}

		// If no PaymentMethod exists, or the existing method is not "stripe", create a new one
		if err == gorm.ErrRecordNotFound || existingPaymentMethod.Method != "stripe" {
			// Step 3: Create a customer in Stripe
			params := &stripe.CustomerParams{
				Email: stripe.String(user.Email), // Set email or other details from the user model
			}
			stripeCustomer, err := customer.New(params)
			if err != nil {
				return err // Return error if Stripe customer creation fails
			}

			// Step 4: Create a new PaymentMethod with Stripe Customer ID
			paymentMethod := PaymentMethod{
				UserID:     user.ID,
				User:       user,
				Method:     "stripe",          // Set the payment method to Stripe
				CustomerID: stripeCustomer.ID, // Store Stripe's Customer ID
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			// Step 5: Save the PaymentMethod to the database
			err = initials.DB.Create(&paymentMethod).Error
			if err != nil {
				return err // Return error if the creation fails
			}

			// Optionally, log the successful creation
			log.Printf("Created PaymentMethod for user: %d with Stripe CustomerID: %s", user.ID, stripeCustomer.ID)
		}
	}

	// If everything is successful, return nil (no errors)
	return nil
} //for test

// FetchStripePlans retrieves the available plans from Stripe
func FetchStripePlans() ([]*stripe.Price, error) {
	// Set your Stripe Secret Key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// List all prices (plans) from Stripe
	params := &stripe.PriceListParams{}
	params.Limit = stripe.Int64(10) // You can adjust the limit as needed

	iter := price.List(params)
	var plans []*stripe.Price

	// Iterate through the plans and append them to the plans slice
	for iter.Next() {
		plans = append(plans, iter.Price())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}
