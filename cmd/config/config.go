package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

func LoadEnv(path ...string) {
	config := &config{}

	if path != nil {
		if err := godotenv.Load(path...); err != nil {
			log.Fatalf("Error loading .env file")
		}
		log.Println("load .env file")
	}

	if err := env.Parse(&config.R2); err != nil {
		log.Fatalf("env load error: %v", err)
	}

	if err := env.Parse(&config.Postgres); err != nil {
		log.Fatalf("env load error: %v", err)
	}

	Config = config
}
