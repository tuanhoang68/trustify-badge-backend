package storage

import (
	"time"

	"gorm.io/gorm"
)

type Shop struct {
	ID string `gorm:"type:char(36);primaryKey"`
	// Thêm size:255 hoặc type:varchar(255)
	ShopDomain  string `gorm:"type:varchar(255);uniqueIndex;not null"`
	AccessToken string `gorm:"not null"`
	InstalledAt time.Time
	UpdatedAt   time.Time

	BadgeConfigs []BadgeConfig `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE"`
}

type BadgeConfig struct {
	ID            string `gorm:"type:char(36);primaryKey"`
	ShopID        string `gorm:"type:char(36);not null;index"`
	Layout        string `gorm:"not null"`
	ReviewText    string `gorm:"type:text"`
	PoweredByText string `gorm:"type:text"`
	HeaderText    string `gorm:"type:text"`
	CustomCode    string `gorm:"type:text"`
	CustomCSS     string `gorm:"type:text"`
	Alignment     string `gorm:"not null;default:center"`
	UpdatedAt     time.Time

	// Quan hệ ngược lại: mỗi config thuộc về một shop
	Shop Shop `gorm:"foreignKey:ShopID;references:ID"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Shop{}, &BadgeConfig{})
}
