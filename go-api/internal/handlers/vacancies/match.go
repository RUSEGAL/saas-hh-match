package handlers_vacancies

import (
	"net/http"
	"strconv"

	"go-api/internal/helpers"
	"go-api/internal/logger"
	"go-api/internal/service/vacancies"

	"github.com/gin-gonic/gin"
)

type MatchRequest struct {
	ResumeID        int64    `json:"resume_id" binding:"required"`
	Query           string   `json:"query" binding:"required"`
	Limit           int      `json:"limit"`
	ExcludeWords    []string `json:"exclude_words"`
	EmploymentTypes []string `json:"employment_types"`
	WorkFormats     []string `json:"work_formats"`
}

type VacancyActionRequest struct {
	VacancyID int64 `json:"vacancy_id" binding:"required"`
}

func MatchVacancies(c *gin.Context) {
	var req MatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	match, err := vacancies.CreateMatchJob(userID, req.ResumeID, req.Query, req.Limit, req.ExcludeWords, req.EmploymentTypes, req.WorkFormats)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to create match job")
		helpers.Respond(c, http.StatusInternalServerError, "job_error", err.Error())
		return
	}

	logger.Info().Int64("match_id", match.ID).Int64("user_id", userID).Str("query", req.Query).Msg("match job created")
	c.JSON(http.StatusOK, gin.H{
		"match_id": match.ID,
		"status":   match.Status,
	})
}

func GetMatchResults(c *gin.Context) {
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

	match, results, err := vacancies.GetMatchWithResults(id)
	if err != nil {
		logger.Error().Err(err).Int64("match_id", id).Msg("failed to get match")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	if match.UserID != userID {
		helpers.Respond(c, http.StatusForbidden, "forbidden", "access denied")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match":   match,
		"results": results,
	})
}

func GetUserMatches(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	matches, err := vacancies.GetUserMatches(userID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Msg("failed to get user matches")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"matches": matches})
}

func SaveVacancyResponse(c *gin.Context) {
	var req VacancyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "bad_request", "user_id required")
		return
	}

	if err := vacancies.SaveResponse(userID, req.VacancyID); err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Int64("vacancy_id", req.VacancyID).Msg("failed to save response")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func SaveVacancyView(c *gin.Context) {
	var req VacancyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.Respond(c, http.StatusBadRequest, "bad_request", "user_id required")
		return
	}

	if err := vacancies.SaveView(userID, req.VacancyID); err != nil {
		logger.Error().Err(err).Int64("user_id", userID).Int64("vacancy_id", req.VacancyID).Msg("failed to save view")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
