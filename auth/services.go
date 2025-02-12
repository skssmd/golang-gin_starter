package auth

import (
	"base/initials"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/gomail.v2"
)

func SyncDatabase() {
	initials.DB.AutoMigrate((&User{}))
}

// GenerateJWT - Generates a JWT for the user
func generateJWT(userID uint, tokenexpiry time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(tokenexpiry).Unix(), // Token expires in 24 hours
	})

	return token.SignedString([]byte(os.Getenv("SECRET")))
}

func sendPasswordResetEmail(toEmail, resetURL string) error {
	// Set up SMTP configuration from environment variables
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("SMTP_USER"))
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", "Password Reset Request")
	mailer.SetBody("text/plain", fmt.Sprintf("Click the following link to reset your password: %s", resetURL))

	// Set up SMTP server details from environment variables
	dialer := gomail.NewDialer(os.Getenv("SMTP_HOST"), 587, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))

	// Send the email
	if err := dialer.DialAndSend(mailer); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Function to send verification email
func sendVerificationEmail(toEmail, verificationURL string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("EMAIL_FROM"))
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", "Email Verification")
	mailer.SetBody("text/plain", fmt.Sprintf("Please verify your email by clicking the following link:\n\n%s", verificationURL))

	dialer := gomail.NewDialer(
		os.Getenv("SMTP_SERVER"),
		587, // Port number for SMTP server
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	if err := dialer.DialAndSend(mailer); err != nil {
		return err
	}

	return nil
}
