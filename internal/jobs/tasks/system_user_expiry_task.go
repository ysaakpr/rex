package tasks

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SystemUserExpiryTask handles the deactivation of expired system user credentials
type SystemUserExpiryTask struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSystemUserExpiryTask(db *gorm.DB, logger *zap.Logger) *SystemUserExpiryTask {
	return &SystemUserExpiryTask{
		db:     db,
		logger: logger,
	}
}

func (t *SystemUserExpiryTask) HandleSystemUserExpiry(ctx context.Context, task *asynq.Task) error {
	t.logger.Info("Starting system user expiry job")

	// Deactivate all expired credentials directly using GORM
	result := t.db.Table("system_users").
		Where("expires_at IS NOT NULL").
		Where("expires_at < NOW()").
		Where("is_active = ?", true).
		Update("is_active", false)

	if result.Error != nil {
		t.logger.Error("Failed to deactivate expired credentials",
			zap.Error(result.Error),
		)
		return fmt.Errorf("failed to deactivate expired credentials: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		t.logger.Info("Deactivated expired credentials",
			zap.Int64("count", result.RowsAffected),
		)
	} else {
		t.logger.Debug("No expired credentials found")
	}

	return nil
}
