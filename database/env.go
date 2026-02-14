package database

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/config"
)

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // File doesn't exist, not an error
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// getDatabaseConfig retrieves database configuration from settings
func getDatabaseConfig() (*config.DatabaseConfig, error) {
	// Load environment variables from file if it exists
	if err := loadEnvFile("/etc/x-ui/db.env"); err != nil {
		log.Printf("Warning: Could not load database environment file: %v", err)
	}

	// Try to get configuration from settings
	// This is a simplified version - in real implementation you'd get this from SettingService
	dbConfig := config.GetDefaultDatabaseConfig()

	// Load configuration from environment variables
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		dbConfig.Type = config.DatabaseType(dbType)
	}

	if dbConfig.Type == config.DatabaseTypePostgreSQL {
		if host := os.Getenv("DB_HOST"); host != "" {
			dbConfig.Postgres.Host = host
		}
		if port := os.Getenv("DB_PORT"); port != "" {
			if p, err := strconv.Atoi(port); err == nil {
				dbConfig.Postgres.Port = p
			}
		}
		if database := os.Getenv("DB_NAME"); database != "" {
			dbConfig.Postgres.Database = database
		}
		if username := os.Getenv("DB_USER"); username != "" {
			dbConfig.Postgres.Username = username
		}
		if password := os.Getenv("DB_PASSWORD"); password != "" {
			dbConfig.Postgres.Password = password
		}
		if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
			dbConfig.Postgres.SSLMode = sslMode
		}
		if timeZone := os.Getenv("DB_TIMEZONE"); timeZone != "" {
			dbConfig.Postgres.TimeZone = timeZone
		}
	}

	return dbConfig, nil
}
