package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint `gorm:"primaryKey"` // internal PK
	CreatedAt time.Time
	UpdatedAt time.Time

	GoogleSub  string `gorm:"size:255;uniqueIndex"` // the "sub"/id from Google
	Email      string `gorm:"size:255"`             // for display only
	Name       string `gorm:"size:255"`             // full name
	PictureURL string `gorm:"size:512"`             // profile picture URL
}

type Color struct {
	gorm.Model
	UserID        uint   `gorm:"not null; index"` // FK to user ID
	Hex           string // Hex value of the color (from user)
	UserColorName string // Name for the color given by user
	RealColorName string // Real name of the color (nearest to be percise)
	Match         bool   // Is it exact match or not
	Palette       string // Color pallete
}
