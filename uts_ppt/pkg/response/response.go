package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API response structure
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginatedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string, errors interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Status:  "error",
		Message: message,
		Errors:  errors,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func Paginated(c *gin.Context, message string, items interface{}, total int64, page, limit int) {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: message,
		Data: PaginatedData{
			Items:      items,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}
