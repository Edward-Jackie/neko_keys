package models

import "time"

// KeyLog records each activate or validate call for a key.
type KeyLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	KeyID     uint      `gorm:"index" json:"key_id"`
	IP        string    `gorm:"type:varchar(64)" json:"ip"`
	Action    string    `gorm:"type:varchar(20)" json:"action"` // activate | validate
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}
