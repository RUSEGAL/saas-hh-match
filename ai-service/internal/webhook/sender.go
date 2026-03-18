package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai-service/internal/ai"
	"ai-service/internal/logger"
	"ai-service/internal/vacancy"
)

type Sender struct {
	client     *http.Client
	url        string
	maxRetries int
}

type WebhookRequest struct {
	ResumeID int64    `json:"resume_id"`
	UserID   int64    `json:"user_id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Score    float64  `json:"score"`
}

type WebhookResponse struct {
	Status string `json:"status"`
}

func NewSender(url string) *Sender {
	return &Sender{
		client:     &http.Client{Timeout: 30 * time.Second},
		url:        url,
		maxRetries: 3,
	}
}

func (s *Sender) Send(resumeID, userID int64, result *ai.AnalysisResult) error {
	webhookReq := WebhookRequest{
		ResumeID: resumeID,
		UserID:   userID,
		Title:    result.OptimizedTitle,
		Content:  result.OptimizedContent,
		Tags:     result.Tags,
		Score:    result.Score,
	}

	jsonBody, err := json.Marshal(webhookReq)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook request: %w", err)
	}

	var lastErr error
	backoff := time.Second

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		req, err := http.NewRequest("POST", s.url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warn().Int("attempt", attempt).Err(err).Msg("Webhook request failed")
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				logger.Info().Int64("resume_id", resumeID).Int("status", resp.StatusCode).Msg("Webhook success")
				return nil
			}

			lastErr = fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
			logger.Warn().Int("attempt", attempt).Int("status", resp.StatusCode).Msg("Webhook returned non-success")
		}

		if attempt < s.maxRetries {
			logger.Info().Dur("backoff", backoff).Msg("Retrying webhook")
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", s.maxRetries, lastErr)
}

type MatchesSender struct {
	client     *http.Client
	url        string
	maxRetries int
}

type MatchPayload struct {
	VacancyTitle   string  `json:"vacancy_title"`
	VacancyCompany string  `json:"vacancy_company"`
	Score          float64 `json:"score"`
	URL            string  `json:"url"`
	Salary         string  `json:"salary"`
	Excerpt        string  `json:"excerpt"`
}

type MatchesWebhookRequest struct {
	UserID   int64          `json:"user_id"`
	ResumeID int64          `json:"resume_id"`
	Query    string         `json:"query"`
	Matches  []MatchPayload `json:"matches"`
}

func NewMatchesSender(url string) *MatchesSender {
	return &MatchesSender{
		client:     &http.Client{Timeout: 60 * time.Second},
		url:        url,
		maxRetries: 3,
	}
}

func (s *MatchesSender) Send(userID, resumeID int64, query string, matches []vacancy.MatchResult) error {
	matchesPayload := make([]MatchPayload, 0, len(matches))
	for _, m := range matches {
		matchesPayload = append(matchesPayload, MatchPayload{
			VacancyTitle:   m.Vacancy.Title,
			VacancyCompany: m.Vacancy.Company,
			Score:          m.Score,
			URL:            m.Vacancy.URL,
			Salary:         m.Vacancy.Salary,
			Excerpt:        m.Vacancy.Excerpt,
		})
	}

	webhookReq := MatchesWebhookRequest{
		UserID:   userID,
		ResumeID: resumeID,
		Query:    query,
		Matches:  matchesPayload,
	}

	jsonBody, err := json.Marshal(webhookReq)
	if err != nil {
		return fmt.Errorf("failed to marshal matches webhook request: %w", err)
	}

	var lastErr error
	backoff := time.Second

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		req, err := http.NewRequest("POST", s.url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create matches webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warn().Int("attempt", attempt).Err(err).Msg("Matches webhook request failed")
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				logger.Info().Int64("resume_id", resumeID).Int("matches", len(matches)).Int("status", resp.StatusCode).Msg("Matches webhook success")
				return nil
			}

			lastErr = fmt.Errorf("matches webhook returned status %d: %s", resp.StatusCode, string(body))
			logger.Warn().Int("attempt", attempt).Int("status", resp.StatusCode).Msg("Matches webhook returned non-success")
		}

		if attempt < s.maxRetries {
			logger.Info().Dur("backoff", backoff).Msg("Retrying matches webhook")
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("matches webhook failed after %d attempts: %w", s.maxRetries, lastErr)
}
