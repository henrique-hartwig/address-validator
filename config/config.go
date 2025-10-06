package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	GeocodingAAPIKey  string
	GeocodingABaseURL string
	GeocodingBAPIKey  string
	GeocodingBBaseURL string
	CacheTTL          time.Duration
	Environment       string
	RedisHost         string
	RedisPort         int
	RedisPassword     string
	RedisDB           int
	APIToken          string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "3000"),
		GeocodingAAPIKey:  getEnv("GEOCODING_A_API_KEY", ""),
		GeocodingABaseURL: getEnv("GEOCODING_A_BASE_URL", "https://api.geoapify.com/v1/geocode/search"),
		GeocodingBAPIKey:  getEnv("GEOCODING_B_API_KEY", ""),
		GeocodingBBaseURL: getEnv("GEOCODING_B_BASE_URL", " https://us-autocomplete-pro.api.smarty.com/lookup"),
		CacheTTL:          parseDuration(getEnv("CACHE_TTL", "24h")),
		Environment:       getEnv("ENVIRONMENT", "development"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         parseInt(getEnv("REDIS_PORT", "6379")),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           parseInt(getEnv("REDIS_DB", "0")),
		APIToken:          getEnv("API_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
