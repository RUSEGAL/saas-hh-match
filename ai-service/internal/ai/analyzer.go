package ai

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

type Analyzer struct {
	client  *http.Client
	apiKey  string
	model   string
	prompt  string
	baseURL string
	cache   *LRUCache
	cacheMu sync.RWMutex
}

type AnalysisResult struct {
	OptimizedTitle   string   `json:"optimized_title"`
	OptimizedContent string   `json:"optimized_content"`
	Tags             []string `json:"tags"`
	Score            float64  `json:"score"`
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type LRUCache struct {
	mu    sync.Mutex
	items map[string]*AnalysisResult
	size  int
	cap   int
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		items: make(map[string]*AnalysisResult),
		cap:   capacity,
	}
}

func (c *LRUCache) Get(key string) (*AnalysisResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.items[key]; ok {
		return val, true
	}
	return nil, false
}

func (c *LRUCache) Set(key string, value *AnalysisResult) {
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

func normalizeTitle(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func NewAnalyzer(apiKey, model, baseURL, promptPath string, cacheSize int) (*Analyzer, error) {
	prompt, err := os.ReadFile(promptPath)
	if err != nil {
		logger.Warn().Err(err).Str("path", promptPath).Msg("Failed to read prompt file, using default")
		prompt = []byte(defaultPrompt)
	}

	if baseURL == "" {
		baseURL = "https://llms.dotpoin.com/v1"
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	if cacheSize <= 0 {
		cacheSize = 1000
	}

	return &Analyzer{
		client:  &http.Client{Timeout: 60 * time.Second},
		apiKey:  apiKey,
		model:   model,
		prompt:  string(prompt),
		baseURL: baseURL,
		cache:   NewLRUCache(cacheSize),
	}, nil
}

func (a *Analyzer) Analyze(title, resumeContent string) (*AnalysisResult, error) {
	normalized := normalizeTitle(title)
	if cached, ok := a.cache.Get(normalized); ok {
		logger.Info().Str("title", title).Msg("Using cached result")
		return cached, nil
	}

	prompt := strings.Replace(a.prompt, "{title}", title, 1)
	prompt = strings.Replace(prompt, "{content}", resumeContent, 1)

	reqBody := OpenAIRequest{
		Model:       a.model,
		Messages:    []Message{{Role: "user", Content: prompt}},
		MaxTokens:   500,
		Temperature: 0,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", a.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	logger.Info().
		Str("url", a.baseURL+"/chat/completions").
		Str("model", a.model).
		Str("body", string(jsonBody)).
		Msg("AI request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	var resp *http.Response
	var lastErr error
	for i := 0; i < 3; i++ {
		resp, lastErr = a.client.Do(req)
		if lastErr == nil {
			break
		}
		logger.Warn().Int("attempt", i+1).Err(lastErr).Msg("OpenAI request failed, retrying")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to call OpenAI after 3 attempts: %w", lastErr)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Info().Int("status", resp.StatusCode).Str("response", string(body)).Msg("AI response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var aiResp OpenAIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := aiResp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		logger.Error().Str("content", content).Err(err).Msg("Failed to parse AI response as JSON")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	a.cache.Set(normalized, &result)
	logger.Info().Str("title", title).Msg("Analysis complete, cached")

	return &result, nil
}

const defaultPrompt = `Analyze this resume title and optimize it.
Title: {title}
Return JSON:
{
  "optimized_title": "...",
  "optimized_content": "...",
  "tags": ["skill1", "skill2"],
  "score": 0.0-1.0
}`
