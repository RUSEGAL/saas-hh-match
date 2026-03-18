package handlers_admin

import (
	"net/http"
	"strconv"

	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/admin"

	"github.com/gin-gonic/gin"
)

// GetUsers godoc
// @Summary Get all users
// @Description Get list of all users (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/users [get]
func GetUsers(c *gin.Context) {
	users, err := admin.GetUsers()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get users")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.Info().Int("count", len(users)).Msg("users retrieved")
	c.JSON(200, gin.H{"users": users})
}

// GetStats godoc
// @Summary Get user statistics
// @Description Get stats for all users or specific user by ID
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/stats [get]
func GetStats(c *gin.Context) {
	if userID := c.Query("user_id"); userID != "" {
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			helpers.Respond(c, http.StatusBadRequest, "invalid_param", "invalid user_id")
			return
		}
		_, stats, err := admin.GetStats(&id)
		if err != nil {
			logger.Error().Err(err).Int64("user_id", id).Msg("failed to get user stats")
			helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		if stats == nil {
			helpers.Respond(c, http.StatusNotFound, "not_found", "user not found")
			return
		}
		logger.Info().Int64("user_id", id).Msg("user stats retrieved")
		c.JSON(200, stats)
		return
	}

	stats, _, err := admin.GetStats(nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get all stats")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.Info().Int("count", len(stats)).Msg("all stats retrieved")
	c.JSON(200, gin.H{"stats": stats})
}

// GetUserResumes godoc
// @Summary Get user resumes
// @Description Get all resumes for a specific user
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/users/{id}/resumes [get]
func GetUserResumes(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_param", "invalid user_id")
		return
	}

	resumes, err := admin.GetResumes(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get user resumes")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.Info().Int64("user_id", userID).Int("count", len(resumes)).Msg("user resumes retrieved")
	c.JSON(200, gin.H{"resumes": resumes})
}

// GetUserPayments godoc
// @Summary Get user payments
// @Description Get all payments for a specific user
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/users/{id}/payments [get]
func GetUserPayments(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_param", "invalid user_id")
		return
	}

	payments, err := admin.GetPayments(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get user payments")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.Info().Int64("user_id", userID).Int("count", len(payments)).Msg("user payments retrieved")
	c.JSON(200, gin.H{"payments": payments})
}
