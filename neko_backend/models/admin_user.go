package models

import "time"

// AdminUser holds credentials for the management console.
type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(64);uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"type:varchar(255)" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
