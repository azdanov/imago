package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Environment string

const (
	Dev  Environment = "dev"
	Prod Environment = "prod"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not set, using fallback: %s", key, fallback)
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not set, using fallback: %d", key, fallback)
		return fallback
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error converting environment variable %s to int: %v, using fallback: %d", key, err, fallback)
		return fallback
	}
	return intValue
}

func getBoolEnv(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not set, using fallback: %t", key, fallback)
		return fallback
	}
	if value == "true" {
		return true
	}
	return false
}

func GetEnvironment(key string, fallback Environment) Environment {
	env := os.Getenv(key)
	switch strings.ToLower(env) {
	case "dev":
		return Dev
	case "prod":
		return Prod
	default:
		log.Printf("Unknown environment '%s', defaulting to %s", env, fallback)
		return fallback
	}
}
