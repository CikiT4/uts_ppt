package middleware

import (
	"strings"
	"time"

	"legal-consultation-api/internal/models"
	jwtpkg "legal-consultation-api/pkg/jwt"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	UserEmailKey = "user_email"
)

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Authorization header format must be: Bearer {token}")
			c.Abort()
			return
		}

		claims, err := jwtpkg.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

// RoleMiddleware restricts access by user role
func RoleMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleKey)
		if !exists {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		userRole := role.(models.UserRole)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "You don't have permission to access this resource")
		c.Abort()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(UserIDKey)
	return id.(uuid.UUID)
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) models.UserRole {
	role, _ := c.Get(UserRoleKey)
	return role.(models.UserRole)
}

// LoggingMiddleware logs every incoming request
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		gin.DefaultWriter.Write([]byte(
			formatLog(start, clientIP, method, path, statusCode, latency),
		))
	}
}

func formatLog(t time.Time, ip, method, path string, status int, latency time.Duration) string {
	return strings.Join([]string{
		"[GIN]",
		t.Format("2006/01/02 - 15:04:05"),
		"|", statusColor(status), "|",
		latency.String(),
		"|", ip,
		"|", method, path, "\n",
	}, " ")
}

func statusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "200"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	default:
		return "5xx"
	}
}

// CORSMiddleware sets CORS headers
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	originMap := make(map[string]bool)
	for _, o := range origins {
		originMap[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if originMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
