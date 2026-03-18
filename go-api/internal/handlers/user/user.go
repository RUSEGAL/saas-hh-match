package handlers_user

import (
	"net/http"

	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/user"

	"github.com/gin-gonic/gin"
)

func GetMyStats(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	stats, err := user.GetUserStats(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get user stats")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetMyPaymentStatus(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	status, err := user.GetPaymentStatus(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get payment status")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}
