package models

import (
	"time"
)

type User struct {
	ID            string    `gorm:"primaryKey;type:text"`
	Name          string    `gorm:"type:text;not null"`
	Email         string    `gorm:"type:text;not null;unique"`
	EmailVerified bool      `gorm:"type:bool;not null;default:false"`
	Image         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "user"
}
