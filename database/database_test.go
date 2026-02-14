package database

import (
	"fmt"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
	"log"
	"testing"

	"github.com/mhsanaei/3x-ui/v2/config"

	"github.com/stretchr/testify/assert"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetDatabaseConnection tests database connection with provided configuration
func GetDatabaseConnection(dbConfig *config.DatabaseConfig) error {
	// Validate configuration
	if err := dbConfig.ValidateConfig(); err != nil {
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

	// Test database connection based on type
	var testDB *gorm.DB
	var err error
	switch dbConfig.Type {
	case config.DatabaseTypeSQLite:
		testDB, err = gorm.Open(sqlite.Open(dbConfig.GetDSN()), c)
	case config.DatabaseTypePostgreSQL:
		testDB, err = gorm.Open(postgres.Open(dbConfig.GetDSN()), c)
	default:
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	if err != nil {
		return err
	}

	// Test the connection
	sqlDB, err := testDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

func initModelsTest(database *gorm.DB) error {
	models := []any{
		&model.User{},
		&model.Inbound{},
		&model.OutboundTraffics{},
		&model.Setting{},
		&model.InboundClientIps{},
		&xray.ClientTraffic{},
		&model.HistoryOfSeeders{},
	}

	for _, dbModel := range models {
		if err := database.AutoMigrate(dbModel); err != nil {
			log.Printf("Error auto migrating model: %v", err)
			return err
		}
	}
	return nil
}

func TestDatabaseConnection(t *testing.T) {
	postgresCfg := &config.DatabaseConfig{
		Type: config.DatabaseTypePostgreSQL,
		SQLite: config.SQLiteConfig{
			Path: "./",
		},
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     8093,
			Database: "test_xui",
			Username: "test_xui",
			Password: "test_xui",
			SSLMode:  "disable",
			TimeZone: "",
		},
	}
	assert.NoError(t, postgresCfg.ValidateConfig())
	assert.NoError(t, postgresCfg.EnsureDirectoryExists())

	assert.NoError(t, GetDatabaseConnection(postgresCfg))

	var gormLogger logger.Interface
	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}

	testDB, err := gorm.Open(postgres.Open(postgresCfg.GetDSN()), c)
	assert.NoError(t, err)

	assert.NoError(t, initModelsTest(testDB))
}
