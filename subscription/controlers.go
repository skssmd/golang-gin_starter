package subscription

import (
	"base/auth"
	"base/initials"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
)

func GetPlans(c *gin.Context) {
	// Initialize the Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Retrieve the plans from Stripe
	params := &stripe.PriceListParams{}
	i := price.List(params) // Get an iterator

	var response []map[string]interface{}

	// Iterate through the results
	for i.Next() {
		plan := i.Price()

		// Fetch product details (Name) using the Product ID from the plan
		prod, err := product.Get(plan.Product.ID, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error fetching product details"})
			return
		}

		// Append plan details including product name to the response
		response = append(response, map[string]interface{}{
			"id":       plan.ID,
			"name":     prod.Name,             // Correct product name
			"price":    plan.UnitAmount / 100, // Assuming Stripe returns the amount in cents
			"currency": plan.Currency,
			"interval": plan.Recurring.Interval,
		})
	}

	// Check for errors after iterating
	if err := i.Err(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(200, response)
}

func CreateSubscription(c *gin.Context) {
	user := auth.GetUser(c) // Get the authenticated user

	// Retrieve the user's selected payment method
	var paymentMethod PaymentMethod
	err := initials.DB.Where("user_id = ?", user.ID).First(&paymentMethod).Error // Fetch payment method from the database
	if err != nil {
		c.JSON(400, gin.H{"error": "No payment method found"})
		return
	}

	// If the payment method is not 'stripe', return an error
	if paymentMethod.Method != "stripe" {
		c.JSON(400, gin.H{"error": "Only Stripe payment method is supported for subscriptions"})
		return
	}

	// Use your Stripe Secret Key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Step 1: Check if the user already has a Stripe Customer ID
	var stripeCustomerID string
	// Retrieve the user's StripeCustomerID from the database or PaymentMethod table
	if paymentMethod.CustomerID == "" {
		// Step 2: Create a new Stripe Customer (if not already created)
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
		}
		stripeCustomer, err := customer.New(customerParams)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error creating Stripe customer"})
			return
		}

		// Save the Stripe Customer ID to the PaymentMethod record
		paymentMethod.CustomerID = stripeCustomer.ID
		// You should save this updated paymentMethod in the database here
		err = initials.DB.Save(&paymentMethod).Error
		if err != nil {
			c.JSON(500, gin.H{"error": "Error saving payment method"})
			return
		}
		stripeCustomerID = stripeCustomer.ID
	} else {
		stripeCustomerID = paymentMethod.CustomerID
	}

	// Step 3: Create the Checkout Session
	sessionParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Customer:           stripe.String(stripeCustomerID), // Use the existing customer ID
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Your Plan Name"), // Dynamically set plan name
					},
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)), // or "year" for yearly plans
					},
					UnitAmount: stripe.Int64(1000), // 1000 cents = $10
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String("https://yourdomain.com/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("https://yourdomain.com/cancel"),
	}

	// Step 4: Create the Checkout session
	session, err := session.New(sessionParams)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Return the session URL to redirect the user to Stripe's checkout page
	c.JSON(200, gin.H{"url": session.URL})
}

func PaymentSuccess(c *gin.Context) {
	sessionID := c.DefaultQuery("session_id", "")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "Session ID is required"})
		return
	}

	// Retrieve the session from Stripe
	session, err := session.Get(sessionID, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Here you can update your database to mark the subscription as active
	// Create a new subscription record in your database, etc.
	c.JSON(200, gin.H{"message": "Payment successful", "session": session})
}
