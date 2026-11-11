package config

import "os"

type Config struct {
	Port   string
	AppEnv string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	
	AccessTokenSecret   string
    RefreshTokenSecret  string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "typetype"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		AccessTokenSecret:  getEnv("ACCESS_TOKEN_SECRET",""),
        RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET",""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}