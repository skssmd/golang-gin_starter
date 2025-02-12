package controlPanel

import (
	"base/auth"
	"base/initials"
	"base/subscription"

	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {
	controlPanel := NewControl(initials.DB)

	// Register Models
	controlPanel.Register("users", auth.User{})
	controlPanel.Register("paymentMethod", subscription.PaymentMethod{})
	controlPanel.Register("Payment", subscription.Payment{})

	SetupControlRoutes(r, controlPanel)
}
