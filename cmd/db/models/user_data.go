package models

import (
	"time"

	"gorm.io/gorm"
)

type UserData struct {
	UserID       string         `gorm:"primaryKey;type:text;not null;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"` // foreign Key referencing User.ID
	SpotifyID    string         `gorm:"unique;not null"`
	DisplayName  string         `gorm:"not null"`
	AccessToken  string         `gorm:"not null"`
	RefreshToken string         `gorm:"not null"`
	LastUpdated  time.Time      `gorm:"autoUpdateTime"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
