package models

import "time"

// Key represents a single license key.
type Key struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	BatchID     uint       `gorm:"index" json:"batch_id"`
	Key         string     `gorm:"type:varchar(64);uniqueIndex" json:"key"`
	Type        string     `gorm:"type:varchar(20)" json:"type"`                    // one_time | multi_use
	Status      string     `gorm:"type:varchar(20);default:inactive" json:"status"` // inactive | active | expired | disabled
	ExpiresAt   time.Time  `json:"expires_at"`
	MaxUses     int        `json:"max_uses"`
	UseCount    int        `gorm:"default:0" json:"use_count"`
	Note        string     `gorm:"type:varchar(255)" json:"note"`
	BoundShop   string     `gorm:"type:varchar(255)" json:"bound_shop"` // shop name bound at first activation (one key = one shop)
	ActivatedAt *time.Time `json:"activated_at"`
	ActivatedIP string     `gorm:"type:varchar(64)" json:"activated_ip"`
	CreatedAt   time.Time  `json:"created_at"`
}

// IsExpired reports whether the key is past its expiration date.
func (k *Key) IsExpired() bool {
	return k.ExpiresAt.Before(time.Now())
}

// IsUsageLimitReached reports whether the key has reached its usage limit.
func (k *Key) IsUsageLimitReached() bool {
	switch k.Type {
	case "one_time":
		return k.UseCount >= 1
	case "multi_use":
		return k.MaxUses > 0 && k.UseCount >= k.MaxUses
	default:
		return false
	}
}

// IsValid reports whether the key can still be used.
func (k *Key) IsValid() bool {
	if k.Status == "disabled" {
		return false
	}
	if k.IsExpired() {
		return false
	}
	if k.IsUsageLimitReached() {
		return false
	}
	return true
}
