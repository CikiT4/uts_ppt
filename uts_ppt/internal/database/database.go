package database

import (
	"database/sql"
	"log"
	"time"

	"legal-consultation-api/internal/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := config.AppConfig.GetDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	DB = db
	log.Println("✅ Database connected successfully")
}

func Close() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
