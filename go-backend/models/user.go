package models

import "time"

type User struct {
	ID           string    `gorm:"primaryKey"`
	Email        string    `gorm:"not null;unique"`
	PasswordHash string    `gorm:"not null"`
	UserName     string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
