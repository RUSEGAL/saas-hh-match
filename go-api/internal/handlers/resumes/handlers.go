package resumes_handlers

import (
	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/resumes"
	types_external "go-api/internal/types/external"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddResume(c *gin.Context) {
	var req types_external.ResumeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if req.Title == "" {
		helpers.Respond(c, http.StatusBadRequest, "validation_error", "title is required")
		return
	}

	userID := req.UserID
	if userID == 0 {
		helpers.Respond(c, http.StatusBadRequest, "bad_request", "user_id is required")
		return
	}

	id, err := resumes.CreateResume(userID, &req)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to create resume")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("user_id", userID).Int64("resume_id", id).Str("title", req.Title).Msg("resume created")
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func UpdateResume(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_id", "wrong id format")
		return
	}

	var req types_external.ResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	userID := req.UserID
	if userID == 0 {
		helpers.Respond(c, http.StatusBadRequest, "bad_request", "user_id is required")
		return
	}

	err = resumes.UpdateResume(id, userID, &req)
	if err != nil {
		logger.Error().Err(err).Int64("resume_id", id).Int64("user_id", userID).Msg("failed to update resume")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("resume_id", id).Int64("user_id", userID).Msg("resume updated")
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// GetMyResumes godoc
// @Summary Get my resumes
// @Description Get all resumes for the authenticated user
// @Tags Resumes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []types_internal.Resume
// @Router /api/resumes/me [get]
func GetMyResumes(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	userResumes, err := resumes.GetResumeByUser(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get resumes")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("user_id", userID).Int("count", len(userResumes)).Msg("resumes retrieved")
	c.JSON(http.StatusOK, userResumes)
}
