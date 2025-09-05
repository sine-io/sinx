package entity

import "time"

// UserRole 用户角色关联
type UserRole struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"index;not null"`
	RoleID    uint      `json:"roleId" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserRole) TableName() string { return "user_roles" }

// RoleMenu 角色菜单关联
type RoleMenu struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RoleID    uint      `json:"roleId" gorm:"index;not null"`
	MenuID    uint      `json:"menuId" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RoleMenu) TableName() string { return "role_menus" }
