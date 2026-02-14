package database

import (
	"fmt"

	"github.com/mhsanaei/3x-ui/v2/config"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDBWithConfig initializes database with provided configuration
func InitDBWithConfig(dbConfig *config.DatabaseConfig) error {
	// Validate configuration
	if err := dbConfig.ValidateConfig(); err != nil {
		return err
	}

	// Ensure directory exists for SQLite
	if err := dbConfig.EnsureDirectoryExists(); err != nil {
		return err
	}

	var gormLogger logger.Interface
	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}

	// Open database connection based on type
	var err error
	switch dbConfig.Type {
	case config.DatabaseTypeSQLite:
		db, err = gorm.Open(sqlite.Open(dbConfig.GetDSN()), c)
	case config.DatabaseTypePostgreSQL:
		db, err = gorm.Open(postgres.Open(dbConfig.GetDSN()), c)
	default:
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	if err != nil {
		return err
	}

	if err := initModels(); err != nil {
		return err
	}

	isUsersEmpty, err := isTableEmpty("users")
	if err != nil {
		return err
	}

	if err := initUser(); err != nil {
		return err
	}
	return runSeeders(isUsersEmpty)
}
