package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ListenAddr  string
	DatabaseDSN string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, falling back to OS environment")
		return nil, err
	}

	return &Config{
		ListenAddr:  getString("LISTEN_ADDR"),
		DatabaseDSN: getString("DATABASE_DSN"),
	}, nil
}

func getString(key string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		fmt.Printf("missing required environment variable: %s\n", key)
	}
	return value
}
