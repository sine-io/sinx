package entity

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单/权限实体
type Menu struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:50;not null"`
	ParentID  uint           `json:"parentId" gorm:"default:0;index"`
	OrderNum  int            `json:"orderNum" gorm:"default:1"`
	Path      string         `json:"path" gorm:"size:100"`
	Component string         `json:"component" gorm:"size:100"`
	Query     string         `json:"query" gorm:"size:100"`
	IsFrame   int16          `json:"isFrame" gorm:"default:0"`
	MenuType  string         `json:"menuType" gorm:"size:2;not null"`
	IsCatch   int16          `json:"isCatch" gorm:"default:0"`
	IsHidden  int16          `json:"isHidden" gorm:"default:0"`
	Perms     string         `json:"perms" gorm:"size:100"`
	Icon      string         `json:"icon" gorm:"size:100"`
	Status    int16          `json:"status" gorm:"default:0"`
	Remark    string         `json:"remark" gorm:"size:100"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Menu) TableName() string { return "menus" }
