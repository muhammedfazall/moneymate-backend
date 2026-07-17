package sharedconfig

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load(path string) {
    if err := godotenv.Load(path); err != nil {
        log.Printf("no .env file found at %s, using environment variables", path)
    }
}

func MustGet(key string) string {
    val := os.Getenv(key)
    if val == "" {
        log.Fatalf("required environment variable %s is not set", key)
    }
    return val
}

func Get(key, fallback string) string {
    val := os.Getenv(key)
    if val == "" {
        return fallback
    }
    return val
}

