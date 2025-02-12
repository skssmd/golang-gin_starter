package initials

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading .env")
	}
}
