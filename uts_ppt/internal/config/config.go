package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort              string
	AppEnv               string
	AppName              string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	JWTSecret            string
	JWTExpireHours       int
	JWTRefreshExpireHours int
	UploadDir            string
	MaxUploadSizeMB      int64
	AllowedOrigins       string
}

var AppConfig *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpire, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	jwtRefreshExpire, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_HOURS", "168"))
	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)

	AppConfig = &Config{
		AppPort:              getEnv("APP_PORT", "8080"),
		AppEnv:               getEnv("APP_ENV", "development"),
		AppName:              getEnv("APP_NAME", "Legal Consultation API"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBName:               getEnv("DB_NAME", "legal_consultation_db"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		JWTSecret:            getEnv("JWT_SECRET", "default_secret_change_in_production"),
		JWTExpireHours:       jwtExpire,
		JWTRefreshExpireHours: jwtRefreshExpire,
		UploadDir:            getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB:      maxUpload,
		AllowedOrigins:       getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
