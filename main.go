package main

import (
	"base/auth"
	"base/controlPanel"
	"base/initials"
	"base/subscription"

	"github.com/gin-gonic/gin"
)

func init() {
	initials.LoadEnvVariables()
	initials.Db()
	auth.SyncDatabase()
	subscription.SyncDatabase()

	// subscription.GenerateStripePaymentMethods() //for test
	println("success")
}

func main() {

	r := gin.Default()
	controlPanel.Routes(r)
	auth.Routes(r)
	subscription.Routes(r)

	r.Run() // Start the server

}
