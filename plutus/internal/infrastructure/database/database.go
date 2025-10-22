package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config contains database configuration parameters
type Config struct {
	DSN string
}

// InitDatabase initializes database connection with configuration
func InitDatabase(config *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established successfully")
	return db, nil
}
