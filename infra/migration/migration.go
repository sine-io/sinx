package migration

import (
	"github.com/sine-io/sinx/domain/user/entity"
	"github.com/sine-io/sinx/pkg/logger"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	logger.Info("Starting database migration...")

	err := db.AutoMigrate(
		&entity.User{},
	)

	if err != nil {
		logger.Error("Database migration failed", "error", err)
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}
