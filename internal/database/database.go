package database

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"skynet-net-engine-api/pkg/logger"
	"go.uber.org/zap"
)

var DB *sql.DB

func Init() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Default fallback for development
		// User provided: fairusinampratama, NO pass
		dsn = "fairusinampratama@tcp(127.0.0.1:3306)/netengine?parseTime=true"
	}

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Fatal("Failed to open database connection")
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Hour)

	if err := DB.Ping(); err != nil {
		logger.Warn("Failed to ping database - Running in Offline/Mock Mode", zap.Error(err))
	} else {
		logger.Info("Database connected successfully")
		Migrate()
	}
}

func Migrate() {
	// 1. Ensure Table Exists (Previous logic assumed it exists, but good to be safe)
	// We skip CREATE TABLE for now as it's likely handled elsewhere or pre-existing
	
	// 2. Add remote_address column if missing
	// We use IGNORE or check approach. Simplest for MySQL is a conditional procedure or just try-catch approach.
	// Since we can't easily do try-catch on DDL in Go without parsing error, we'll try to add it.
	// Duplicate column error is fine to ignore if we check strictly, but here we'll just run generic ADD COLUMN IF NOT EXISTS logic
	// MySQL 8.0 support IF NOT EXISTS in ADD COLUMN, but MariaDB might not in all versions.
	// We will simply try to query it first.
	
	query := "SHOW COLUMNS FROM pppoe_users LIKE 'remote_address'"
	rows, err := DB.Query(query)
	if err != nil {
		logger.Error("Failed to check schema", zap.Error(err))
		return
	}
	defer rows.Close()
	
	if !rows.Next() {
		logger.Info("Migrating DB: Adding remote_address to pppoe_users")
		_, err := DB.Exec("ALTER TABLE pppoe_users ADD COLUMN remote_address VARCHAR(45) DEFAULT NULL")
		if err != nil {
			logger.Error("Failed to migrate database", zap.Error(err))
		}
	}
	rows.Close() // Ensure closed before next query

	// 3. Add is_isolated column
	queryIso := "SHOW COLUMNS FROM pppoe_users LIKE 'is_isolated'"
	rowsIso, errIso := DB.Query(queryIso)
	if errIso != nil {
		logger.Error("Failed to check schema for is_isolated", zap.Error(errIso))
		return
	}
	defer rowsIso.Close()

	if !rowsIso.Next() {
		logger.Info("Migrating DB: Adding is_isolated to pppoe_users")
		_, err := DB.Exec("ALTER TABLE pppoe_users ADD COLUMN is_isolated BOOLEAN DEFAULT FALSE")
		if err != nil {
			logger.Error("Failed to add is_isolated column", zap.Error(err))
		}
	}
}
