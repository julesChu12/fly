package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDatabase initializes database connection
func InitDatabase(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
