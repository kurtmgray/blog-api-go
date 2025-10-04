package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	MongoDB string
	JWTSecret string
	Port string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		MongoDB: os.Getenv("MONGO_DB"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port: os.Getenv("PORT"),
	}
}