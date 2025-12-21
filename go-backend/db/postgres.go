package db

import (
	"fmt"
	"mcpdocs/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(config *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDb)
	return gorm.Open(postgres.Open(dsn))
}
