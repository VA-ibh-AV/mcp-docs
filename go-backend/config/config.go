package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDb       string
	KafkaBrokers     []string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Error("Warning: .env file not found, using environment variables from system", "error", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		slog.Error("PORT is not set")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		slog.Error("POSTGRES_HOST is not set")
	}

	postgresPort := os.Getenv("POSTGRES_PORT")
	if postgresPort == "" {
		slog.Error("POSTGRES_PORT is not set")
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		slog.Error("POSTGRES_USER is not set")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		slog.Error("POSTGRES_PASSWORD is not set")
	}

	postgresDb := os.Getenv("POSTGRES_DB")
	if postgresDb == "" {
		slog.Error("POSTGRES_DB is not set")
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	kafkaBrokers := []string{"kafka:29092"} // Default value
	if kafkaBrokersStr != "" {
		kafkaBrokers = []string{kafkaBrokersStr}
	}

	return &Config{
		Port:             port,
		PostgresHost:     postgresHost,
		PostgresPort:     postgresPort,
		PostgresUser:     postgresUser,
		PostgresPassword: postgresPassword,
		PostgresDb:       postgresDb,
		KafkaBrokers:     kafkaBrokers,
	}
}

func GetJWTSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		slog.Error("JWT_SECRET_KEY is not set")
	}
	return secretKey
}
