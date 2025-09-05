package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Username          string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Password          string         `json:"-" gorm:"not null;size:255"`
	Avatar            string         `json:"avatar" gorm:"size:255"`
	Nickname          string         `json:"nickname" gorm:"size:50"`
	UserType          int16          `json:"userType" gorm:"default:0"` // 0 普通 1 超管
	Email             string         `json:"email" gorm:"size:100"`
	Mobile            string         `json:"mobile" gorm:"size:30"`
	Sort              int            `json:"sort" gorm:"default:1"`
	Status            int16          `json:"status" gorm:"default:0"` // 0 正常 1 禁用
	LastLoginIP       string         `json:"lastLoginIp" gorm:"size:30"`
	LastLoginNation   string         `json:"lastLoginNation" gorm:"size:100"`
	LastLoginProvince string         `json:"lastLoginProvince" gorm:"size:100"`
	LastLoginCity     string         `json:"lastLoginCity" gorm:"size:100"`
	LastLoginDate     *time.Time     `json:"lastLoginDate"`
	Salt              string         `json:"-" gorm:"size:30"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string { return "users" }
