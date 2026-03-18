package handlers_ai

import (
	"go-api/internal/helpers"
	"go-api/internal/logger"
	dbresumes "go-api/internal/repository/resumes"
	"go-api/internal/service/vacancies"
	types_internal "go-api/internal/types/int"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResumeAnalysisResult struct {
	ResumeID int64    `json:"resume_id" binding:"required"`
	UserID   int64    `json:"user_id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Score    float64  `json:"score"`
}

func WebhookAnalyzeResume(c *gin.Context) {
	var req ResumeAnalysisResult
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if req.ResumeID <= 0 {
		helpers.Respond(c, http.StatusBadRequest, "validation_error", "resume_id is required")
		return
	}

	err := dbresumes.UpdateResumeContent(req.ResumeID, req.Title, req.Content, req.Tags, req.Score)
	if err != nil {
		logger.Error().Err(err).Int64("resume_id", req.ResumeID).Msg("failed to update resume from AI")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("resume_id", req.ResumeID).Float64("score", req.Score).Msg("resume updated from AI webhook")
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

type VacancyWebhookResult struct {
	UserID   int64                                    `json:"user_id" binding:"required"`
	ResumeID int64                                    `json:"resume_id"`
	MatchID  int64                                    `json:"match_id" binding:"required"`
	Query    string                                   `json:"query"`
	Matches  []types_internal.VacancyMatchResultInput `json:"matches"`
}

func WebhookVacancyMatches(c *gin.Context) {
	var req VacancyWebhookResult
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.Respond(c, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if req.MatchID <= 0 {
		helpers.Respond(c, http.StatusBadRequest, "validation_error", "match_id is required")
		return
	}

	err := vacancies.ProcessWebhookResult(&types_internal.VacancyWebhookResult{
		UserID:   req.UserID,
		ResumeID: req.ResumeID,
		MatchID:  req.MatchID,
		Query:    req.Query,
		Matches:  req.Matches,
	})
	if err != nil {
		logger.Error().Err(err).Int64("match_id", req.MatchID).Msg("failed to process vacancy matches webhook")
		helpers.Respond(c, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	logger.Info().Int64("match_id", req.MatchID).Int("count", len(req.Matches)).Msg("vacancy matches processed from AI webhook")
	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
