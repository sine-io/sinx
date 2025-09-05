package migration

import (
	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	rbacEntity "github.com/sine-io/sinx/domain/rbac/entity"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	userEntity "github.com/sine-io/sinx/domain/user/entity"
	"github.com/sine-io/sinx/pkg/logger"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	logger.Info("Starting database migration...")

	err := db.AutoMigrate(
		&userEntity.User{},
		&roleEntity.Role{},
		&menuEntity.Menu{},
		&rbacEntity.UserRole{},
		&rbacEntity.RoleMenu{},
	)

	if err != nil {
		logger.Error("Database migration failed", "error", err)
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}
