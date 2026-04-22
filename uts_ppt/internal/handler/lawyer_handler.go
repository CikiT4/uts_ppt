package handler

import (
	"strconv"

	"legal-consultation-api/internal/middleware"
	"legal-consultation-api/internal/repository"
	"legal-consultation-api/internal/service"
	"legal-consultation-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LawyerHandler struct {
	lawyerService   service.LawyerService
	scheduleService service.ScheduleService
	reviewService   service.ReviewService
}

func NewLawyerHandler(ls service.LawyerService, ss service.ScheduleService, rs service.ReviewService) *LawyerHandler {
	return &LawyerHandler{lawyerService: ls, scheduleService: ss, reviewService: rs}
}

// GET /api/lawyers — Search with filters
func (h *LawyerHandler) SearchLawyers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	minFee, _ := strconv.ParseFloat(c.Query("min_fee"), 64)
	maxFee, _ := strconv.ParseFloat(c.Query("max_fee"), 64)
	minRating, _ := strconv.ParseFloat(c.Query("min_rating"), 64)

	filter := repository.LawyerFilter{
		Specialization: c.Query("specialization"),
		City:           c.Query("city"),
		MinRating:      minRating,
		MinFee:         minFee,
		MaxFee:         maxFee,
		Search:         c.Query("search"),
		Page:           page,
		Limit:          limit,
	}

	lawyers, total, err := h.lawyerService.SearchLawyers(filter)
	if err != nil {
		response.InternalServerError(c, "Failed to search lawyers")
		return
	}
	response.Paginated(c, "Lawyers retrieved", lawyers, total, page, limit)
}

// GET /api/lawyers/:id
func (h *LawyerHandler) GetLawyer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid lawyer ID", nil)
		return
	}
	lawyer, err := h.lawyerService.GetProfile(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, "Lawyer retrieved", lawyer)
}

// POST /api/lawyers/profile — Lawyer creates own profile
func (h *LawyerHandler) CreateProfile(c *gin.Context) {
	var req service.LawyerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	lawyer, err := h.lawyerService.CreateProfile(userID, &req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, "Lawyer profile created", lawyer)
}

// PUT /api/lawyers/profile
func (h *LawyerHandler) UpdateProfile(c *gin.Context) {
	var req service.LawyerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	lawyer, err := h.lawyerService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	updated, err := h.lawyerService.UpdateProfile(lawyer.ID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Profile updated", updated)
}

// PATCH /api/lawyers/availability
func (h *LawyerHandler) SetAvailability(c *gin.Context) {
	var body struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	lawyer, err := h.lawyerService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if err := h.lawyerService.SetAvailability(lawyer.ID, body.IsAvailable); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Availability updated", nil)
}

// GET /api/lawyers/:id/schedules
func (h *LawyerHandler) GetSchedules(c *gin.Context) {
	lawyerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid lawyer ID", nil)
		return
	}
	schedules, err := h.scheduleService.GetByLawyer(lawyerID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Schedules retrieved", schedules)
}

// POST /api/schedules
func (h *LawyerHandler) CreateSchedule(c *gin.Context) {
	var req service.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	schedule, err := h.scheduleService.Create(userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, "Schedule created", schedule)
}

// DELETE /api/schedules/:id
func (h *LawyerHandler) DeleteSchedule(c *gin.Context) {
	scheduleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid schedule ID", nil)
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.scheduleService.Delete(scheduleID, userID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, "Schedule deleted", nil)
}

// GET /api/lawyers/:id/reviews
func (h *LawyerHandler) GetReviews(c *gin.Context) {
	lawyerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid lawyer ID", nil)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	reviews, total, err := h.reviewService.GetByLawyerID(lawyerID, page, limit)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Paginated(c, "Reviews retrieved", reviews, total, page, limit)
}
