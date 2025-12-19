package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ShopDomain  string    `gorm:"uniqueIndex;not null"` // e.g., demo-store.myshopify.com
	AccessToken string    `gorm:"not null"`
	InstalledAt time.Time
	UpdatedAt   time.Time
}

type BadgeConfig struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ShopDomain    string    `gorm:"index;not null"`
	Layout        string    `gorm:"not null"` // layout1..layout5 or custom
	ReviewText    string    `gorm:"type:text"`
	PoweredByText string    `gorm:"type:text"`
	HeaderText    string    `gorm:"type:text"` // layout5 only
	CustomCode    string    `gorm:"type:text"` // custom layout
	CustomCSS     string    `gorm:"type:text"`
	Alignment     string    `gorm:"not null;default:center"`
	UpdatedAt     time.Time
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Shop{}, &BadgeConfig{})
}
