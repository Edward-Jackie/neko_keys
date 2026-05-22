package models

import "time"

// KeyBatch represents a batch of generated keys.
type KeyBatch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Note      string    `gorm:"type:varchar(255)" json:"note"`
	Count     int       `json:"count"`
	Type      string    `gorm:"type:varchar(20)" json:"type"` // one_time | multi_use
	ExpiresAt time.Time `json:"expires_at"`
	MaxUses   int       `json:"max_uses"`
	CreatedAt time.Time `json:"created_at"`
}
