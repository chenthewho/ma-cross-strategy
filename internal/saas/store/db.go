// Package store provides database connection and AutoMigrate.
package store

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/config"
)

// DB wraps the GORM database instance.
type DB struct {
	*gorm.DB
}

// NewDB creates a new PostgreSQL connection and runs AutoMigrate.
func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	dsn := cfg.DSN()

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	log.Println("[DB] Connected to PostgreSQL, running AutoMigrate...")

	if err := AutoMigrateAll(db); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	log.Println("[DB] AutoMigrate complete")

	return &DB{DB: db}, nil
}
