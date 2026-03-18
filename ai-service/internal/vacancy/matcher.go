package vacancy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"ai-service/internal/logger"
)

type Matcher struct {
	client    *http.Client
	apiKey    string
	model     string
	prompt    string
	baseURL   string
	threshold float64
	batchSize int
	cache     *MatchCache
	cacheMu   sync.RWMutex
}

type MatchResult struct {
	Vacancy   Vacancy
	Score     float64
	Reasoning string
}

type MatchCache struct {
	mu    sync.Mutex
	items map[string][]MatchResult
	size  int
	cap   int
}

func NewMatchCache(capacity int) *MatchCache {
	return &MatchCache{
		items: make(map[string][]MatchResult),
		cap:   capacity,
	}
}

func (c *MatchCache) Get(key string) ([]MatchResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.items[key]; ok {
		return val, true
	}
	return nil, false
}

func (c *MatchCache) Set(key string, value []MatchResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.cap {
		for k := range c.items {
			delete(c.items, k)
			break
		}
	}
	c.items[key] = value
}

func BuildMatchCacheKey(resumeID int64, query string, employmentTypes, workFormats []string, excludeWords []string) string {
	return fmt.Sprintf("%d:%s:%v:%v:%v", resumeID, query, employmentTypes, workFormats, excludeWords)
}

type BatchMatchRequest struct {
	Resume    string    `json:"resume"`
	Vacancies []Vacancy `json:"vacancies"`
}

type VacancyMatchItem struct {
	Title   string `json:"title"`
	Company string `json:"company"`
	Excerpt string `json:"excerpt"`
}

type BatchMatchResponse struct {
	Matches []MatchItem `json:"matches"`
}

type MatchItem struct {
	Index     int     `json:"index"`
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning"`
}

func NewMatcher(apiKey, model, baseURL, promptPath string, threshold float64, batchSize int) (*Matcher, error) {
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		logger.Warn().Err(err).Str("path", promptPath).Msg("Failed to read match prompt file, using default")
		prompt = []byte(defaultBatchMatchPrompt)
	}

	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	if batchSize <= 0 {
		batchSize = 20
	}

	return &Matcher{
		client:    &http.Client{Timeout: 120 * time.Second},
		apiKey:    apiKey,
		model:     model,
		prompt:    string(prompt),
		baseURL:   baseURL,
		threshold: threshold,
		batchSize: batchSize,
		cache:     NewMatchCache(500),
	}, nil
}

func (m *Matcher) MatchVacancies(resumeContent string, vacancies []Vacancy) ([]MatchResult, error) {
	return m.matchVacanciesWithKey(resumeContent, vacancies, "")
}

func (m *Matcher) MatchVacanciesCached(resumeContent string, vacancies []Vacancy, cacheKey string) ([]MatchResult, error) {
	if cacheKey != "" {
		if cached, ok := m.cache.Get(cacheKey); ok {
			logger.Info().Str("cacheKey", cacheKey).Int("count", len(cached)).Msg("Using cached vacancy matches")
			return cached, nil
		}
	}

	results, err := m.matchVacanciesWithKey(resumeContent, vacancies, cacheKey)
	if err != nil {
		return results, err
	}

	if cacheKey != "" {
		m.cache.Set(cacheKey, results)
		logger.Info().Str("cacheKey", cacheKey).Int("count", len(results)).Msg("Cached vacancy matches")
	}

	return results, nil
}

func (m *Matcher) matchVacanciesWithKey(resumeContent string, vacancies []Vacancy, cacheKey string) ([]MatchResult, error) {
	results := make([]MatchResult, 0)

	for i := 0; i < len(vacancies); i += m.batchSize {
		end := i + m.batchSize
		if end > len(vacancies) {
			end = len(vacancies)
		}

		batch := vacancies[i:end]
		logger.Info().Int("batch", i/m.batchSize+1).Int("vacancies", len(batch)).Msg("Processing batch")

		batchResults, err := m.matchBatch(resumeContent, batch)
		if err != nil {
			logger.Error().Err(err).Msg("Batch matching failed, skipping batch")
			continue
		}

		results = append(results, batchResults...)
	}

	filtered := make([]MatchResult, 0)
	for _, r := range results {
		if r.Score >= m.threshold {
			filtered = append(filtered, r)
		}
	}

	logger.Info().Int("total", len(vacancies)).Int("matched", len(filtered)).Float64("threshold", m.threshold).Msg("Vacancy matching complete")

	return filtered, nil
}

func (m *Matcher) matchBatch(resumeContent string, vacancies []Vacancy) ([]MatchResult, error) {
	prompt := m.buildBatchPrompt(resumeContent, vacancies)

	reqBody := map[string]interface{}{
		"model": m.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  2000,
		"temperature": 0,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", m.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	var resp *http.Response
	var lastErr error
	for i := 0; i < 3; i++ {
		resp, lastErr = m.client.Do(req)
		if lastErr == nil {
			break
		}
		logger.Warn().Int("attempt", i+1).Err(lastErr).Msg("AI batch match request failed, retrying")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to call AI after 3 attempts: %w", lastErr)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API error: %s", string(body))
	}

	var aiResp map[string]interface{}
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	choices, ok := aiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract content")
	}

	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var batchResp BatchMatchResponse
	if err := json.Unmarshal([]byte(content), &batchResp); err != nil {
		logger.Error().Str("content", content).Err(err).Msg("Failed to parse batch match response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	results := make([]MatchResult, 0, len(batchResp.Matches))
	for _, match := range batchResp.Matches {
		if match.Index >= 0 && match.Index < len(vacancies) {
			results = append(results, MatchResult{
				Vacancy:   vacancies[match.Index],
				Score:     match.Score,
				Reasoning: match.Reasoning,
			})
		}
	}

	return results, nil
}

func (m *Matcher) buildBatchPrompt(resumeContent string, vacancies []Vacancy) string {
	prompt := m.prompt
	prompt = strings.Replace(prompt, "{resume}", resumeContent, 1)

	vacanciesList := ""
	for i, v := range vacancies {
		vacanciesList += fmt.Sprintf("%d. %s at %s\n   %s\n", i+1, v.Title, v.Company, v.Excerpt)
	}
	prompt = strings.Replace(prompt, "{vacancies}", vacanciesList, 1)

	return prompt
}

const defaultBatchMatchPrompt = `Compare this resume with each vacancy listed below. Return match scores for ALL vacancies.

Resume:
{resume}

Vacancies:
{vacancies}

Return a JSON array with match results for EACH vacancy (in the same order):
[
  {"index": 0, "score": 0.0-1.0, "reasoning": "brief explanation"},
  {"index": 1, "score": 0.0-1.0, "reasoning": "brief explanation"},
  ...
]

Important:
- Return results for ALL vacancies in the list
- index must match the vacancy position (0-based)
- Be concise but accurate in reasoning`
