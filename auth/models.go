package auth

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint   `gorm:"primaryKey"`
	Username    string `gorm:"unique"`
	Email       string `gorm:"unique"`
	Password    string
	FirstName   string `gorm:"default:null"` // Added
	LastName    string `gorm:"default:null"` // Added
	IsVerified  bool   `gorm:"default:false"`
	IsAdmin     bool   `gorm:"default:false"`
	IsSuperuser bool   `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
