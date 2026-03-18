package resumes_handlers

import (
	"net/http"
	"strconv"

	"go-api/internal/helpers"
	"go-api/internal/logger"
	dbresumes "go-api/internal/repository/resumes"

	"github.com/gin-gonic/gin"
)

func DeleteResume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_id", "wrong id format")
		return
	}

	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	err = dbresumes.DeleteResume(id, userID)
	if err != nil {
		logger.Error().Err(err).Int64("resume_id", id).Int64("user_id", userID).Msg("failed to delete resume")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("resume_id", id).Int64("user_id", userID).Msg("resume deleted")
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
