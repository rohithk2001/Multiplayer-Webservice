package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfigStruct struct {
	MongoDBURI string
	RedisAddr  string
	RedisPass  string
	RedisDB    int
	ServerPort string
	GRPCPort   string
}

// AppConfig holds the application configuration
var AppConfig AppConfigStruct

// LoadConfig loads environment variables and validates them
func LoadConfig() error {
    // Load .env file
    err := godotenv.Load("/app/.env")
    if err != nil {
        log.Printf("Warning: .env file not loaded, falling back to environment variables.")
    }

    // Read environment variables
    AppConfig.MongoDBURI = os.Getenv("MONGODB_URI")
    AppConfig.RedisAddr = os.Getenv("REDIS_ADDR")
    AppConfig.RedisPass = os.Getenv("REDIS_PASS")
    AppConfig.RedisDB = getEnvInt("REDIS_DB", 0)
    AppConfig.ServerPort = os.Getenv("SERVER_PORT")
    AppConfig.GRPCPort = os.Getenv("GRPC_PORT")

    // Log configuration
    log.Printf("Loaded configuration: %+v", AppConfig)

    // Validate required fields
    if AppConfig.MongoDBURI == "" || AppConfig.RedisAddr == "" {
        return fmt.Errorf("missing essential environment variables: MONGODB_URI, REDIS_ADDR")
    }

    return nil
}

// getEnv fetches a string environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt fetches an integer environment variable with a fallback default
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("invalid integer value for %s, falling back to default: %d", key, defaultValue)
	}
	return defaultValue
}
