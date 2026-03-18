package resumes_handlers

import (
	"net/http"
	"strconv"

	"go-api/internal/helpers"
	"go-api/internal/logger"
	dbresumes "go-api/internal/repository/resumes"

	"github.com/gin-gonic/gin"
)

func GetResumeByID(c *gin.Context) {
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

	resume, err := dbresumes.GetResumeByID(id)
	if err != nil {
		logger.Error().Err(err).Int64("resume_id", id).Msg("failed to get resume")
		helpers.Respond(c, http.StatusNotFound, "not_found", "resume not found")
		return
	}

	if resume.UserID != userID {
		helpers.Respond(c, http.StatusForbidden, "forbidden", "access denied")
		return
	}

	c.JSON(http.StatusOK, resume)
}
