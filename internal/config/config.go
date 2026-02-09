package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	MongoDBURI        string
	MongoDBDatabase   string
	JWTSecret         string
	JWTAccessExpiry   time.Duration
	JWTRefreshExpiry  time.Duration
	Argon2Memory      uint32
	Argon2Iterations  uint32
	Argon2Parallelism uint8
	Argon2SaltLength  uint32
	Argon2KeyLength   uint32
	LogLevel          string
	Environment       string
	CookieDomain      string
	CookieSecure      bool
	CookieSameSite    string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8085"),
		MongoDBURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:   getEnv("MONGODB_DATABASE", "infrantery"),
		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-key"),
		JWTAccessExpiry:   parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry:  parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		Argon2Memory:      parseUint32(getEnv("ARGON2_MEMORY", "65536")),
		Argon2Iterations:  parseUint32(getEnv("ARGON2_ITERATIONS", "3")),
		Argon2Parallelism: parseUint8(getEnv("ARGON2_PARALLELISM", "2")),
		Argon2SaltLength:  parseUint32(getEnv("ARGON2_SALT_LENGTH", "16")),
		Argon2KeyLength:   parseUint32(getEnv("ARGON2_KEY_LENGTH", "32")),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		CookieDomain:      getEnv("COOKIE_DOMAIN", "localhost"),
		CookieSecure:      getEnv("COOKIE_SECURE", "false") == "true",
		CookieSameSite:    getEnv("COOKIE_SAMESITE", "lax"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

func parseUint32(s string) uint32 {
	val, _ := strconv.ParseUint(s, 10, 32)
	return uint32(val)
}

func parseUint8(s string) uint8 {
	val, _ := strconv.ParseUint(s, 10, 8)
	return uint8(val)
}
