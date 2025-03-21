package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sunriseex/test_wallet/internal/logger"
)

type Config struct {
	AppPort string
	DBHost  string
	DBPort  string
	DBUser  string
	DBPass  string
	DBName  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Error("Error loading config no .env file")
	}
	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		DBHost:  os.Getenv("DB_HOST"),
		DBPort:  os.Getenv("DB_PORT"),
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASS"),
		DBName:  os.Getenv("DB_NAME"),
	}
}
