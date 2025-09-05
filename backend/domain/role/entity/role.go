package entity

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色实体
type Role struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Remark    string         `json:"remark" gorm:"size:100"`
	Status    int16          `json:"status" gorm:"default:0"` // 0正常 1禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Role) TableName() string { return "roles" }
