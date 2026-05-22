package database

import (
	"fmt"
	"log"

	"neko_backend/config"
	"neko_backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the shared database connection.
var DB *gorm.DB

// Init opens the MySQL connection, runs AutoMigrate, and seeds the default admin.
func Init(cfg *config.Config) error {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if err := seedAdmin(cfg); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	return nil
}

func migrate() error {
	return DB.AutoMigrate(
		&models.KeyBatch{},
		&models.Key{},
		&models.KeyLog{},
		&models.AdminUser{},
	)
}

func seedAdmin(cfg *config.Config) error {
	var count int64
	DB.Model(&models.AdminUser{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := models.AdminUser{
		Username:     cfg.AdminUsername,
		PasswordHash: string(hash),
	}
	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Printf("[init] created default admin user: %s", cfg.AdminUsername)
	return nil
}
