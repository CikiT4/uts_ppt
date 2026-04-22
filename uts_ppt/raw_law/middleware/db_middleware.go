package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DBHealthMiddleware ensures the database connection is alive before
// processing any request. Uses a short context timeout to avoid blocking.
func DBHealthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection unavailable",
				"error":   err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database ping failed",
				"error":   err.Error(),
			})
			return
		}

		c.Next()
	}
}

// InjectDB stores *gorm.DB in the Gin context so handlers can retrieve it
// via dependency injection without using global variables.
func InjectDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// GetDB retrieves *gorm.DB from the Gin context.
// Handlers should call this instead of accessing a global variable.
func GetDB(c *gin.Context) *gorm.DB {
	db, exists := c.Get("db")
	if !exists {
		return nil
	}
	return db.(*gorm.DB)
}
