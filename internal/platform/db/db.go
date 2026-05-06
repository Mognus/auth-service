package db

import (
	"fmt"

	"auth-service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(GormDSN(cfg)), &gorm.Config{})
}

func GormDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port,
		cfg.SSLMode,
	)
}

func MigrateDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
}
