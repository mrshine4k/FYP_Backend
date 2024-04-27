package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// basis for loading environment variables
func loadEnv() {
	// if in production, don't load .env file and use the environment variables provided by the server
	if os.Getenv("PRODUCTION") == "TRUE" {
		return
	}
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func EnvMongoURI() string {
	loadEnv()
	return os.Getenv("MONGOURI")
}

func EnvDBName() string {
	loadEnv()
	return os.Getenv("DBNAME")
}
