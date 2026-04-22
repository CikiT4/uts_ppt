package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"raw-law-api/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds MySQL connection parameters
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

// DefaultDBConfig returns the raw_law database config
func DefaultDBConfig() DBConfig {
	return DBConfig{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "",
		DBName:   "raw_law",
		Charset:  "utf8mb4",
	}
}

// DSN builds the MySQL connection string
// Format: root:@tcp(127.0.0.1:3306)/raw_law?parseTime=true&charset=utf8mb4&loc=Local
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.Charset,
	)
}

// InitDB opens a GORM connection to MySQL, configures the connection pool,
// and auto-migrates all registered models. Returns *gorm.DB for DI.
func InitDB(cfg DBConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		DisableForeignKeyConstraintWhenMigrating: true, // Prevents foreign key constraint errors
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get the underlying *sql.DB to configure the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)                // Maximum open connections
	sqlDB.SetMaxIdleConns(5)                 // Maximum idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Max connection lifetime
	sqlDB.SetConnMaxIdleTime(1 * time.Minute) // Max idle connection lifetime

	// Verify connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Auto migrate all models
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	log.Println("✅ MySQL connected and migrated successfully")
	log.Printf("   DSN: %s@tcp(%s:%s)/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)
	return db, nil
}

// autoMigrate runs GORM AutoMigrate for all tables
func autoMigrate(db *gorm.DB) error {
	// Disable foreign key checks temporarily
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// Drop existing foreign keys that might block column modifications
	db.Exec("ALTER TABLE lawyers DROP FOREIGN KEY lawyers_ibfk_1")
	db.Exec("ALTER TABLE lawyers DROP FOREIGN KEY fk_lawyers_user")
	db.Exec("ALTER TABLE consultations DROP FOREIGN KEY consultations_ibfk_1")
	db.Exec("ALTER TABLE consultations DROP FOREIGN KEY consultations_ibfk_2")
	db.Exec("ALTER TABLE payments DROP FOREIGN KEY payments_ibfk_1")
	db.Exec("ALTER TABLE reviews DROP FOREIGN KEY reviews_ibfk_1")
	db.Exec("ALTER TABLE documents DROP FOREIGN KEY documents_ibfk_1")

	return db.AutoMigrate(
		&models.User{},
		&models.Lawyer{},
		&models.Consultation{},
		&models.Payment{},
		&models.Review{},
		&models.Document{},
	)
}
