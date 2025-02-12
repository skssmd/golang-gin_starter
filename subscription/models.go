package subscription

import (
	"base/auth"
	"base/initials"
	"time"
)

type PaymentMethod struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint `gorm:"unique;not null"` // Foreign key to User model
	User       auth.User
	Method     string `gorm:"default:'stripe'"` // Enum type with 'stripe' as the default
	CustomerID string `gorm:"unique"`           // Unified customer ID field (stripe, bkash, nagad)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
type Subscription struct {
	ID                   uint      `gorm:"primaryKey"`                                     // Unique ID (Primary Key)
	UserID               uint      `gorm:"not null;index"`                                 // User ID (Foreign Key to User model)
	StripeSubscriptionID string    `gorm:"unique;not null"`                                // Stripe Subscription ID
	Status               string    `gorm:"not null"`                                       // e.g., "active", "canceled", etc.
	PlanID               string    `gorm:"not null"`                                       // Stripe Plan ID
	StartDate            time.Time `gorm:"not null"`                                       // Subscription start date
	EndDate              time.Time `gorm:"not null"`                                       // Subscription end date
	TrialEndDate         time.Time `gorm:"not null"`                                       // Trial end date (if applicable)
	PaymentMethodID      string    `gorm:"not null"`                                       // Stripe Payment Method ID
	CreatedAt            time.Time `gorm:"autoCreateTime"`                                 // Automatically sets when the subscription is created
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`                                 // Automatically updates when the subscription is modified
	User                 auth.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"` // Association with User model
}

// Ensure to include this in the package where you define the User model, so it has a consistent relationship

type Payment struct {
	ID             uint         `gorm:"primaryKey"`                                             // Unique ID (Primary Key)
	SubscriptionID uint         `gorm:"not null;index"`                                         // Foreign Key to Subscription model
	Amount         float64      `gorm:"not null"`                                               // Payment amount
	Status         string       `gorm:"not null"`                                               // Payment status (e.g., "successful", "failed")
	PaymentMethod  string       `gorm:"not null"`                                               // Payment method (e.g., "Stripe", "Bkash", etc.)
	TransactionID  string       `gorm:"not null"`                                               // Transaction ID from payment gateway
	PaymentDate    time.Time    `gorm:"not null"`                                               // Payment date
	CreatedAt      time.Time    `gorm:"autoCreateTime"`                                         // Automatically sets when the payment is created
	UpdatedAt      time.Time    `gorm:"autoUpdateTime"`                                         // Automatically updates when the payment is modified
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE;"` // Association with Subscription model
}

func SyncDatabase() {
	// Auto migrate Subscription and Payment models
	initials.DB.AutoMigrate(&Subscription{}, &Payment{}, &PaymentMethod{})
}
