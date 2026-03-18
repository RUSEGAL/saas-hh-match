package vacancy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"ai-service/internal/logger"
)

type Fetcher struct {
	client      *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	cache       *VacancyCache
}

type HHVacancy struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Employer     Employer `json:"employer"`
	Salary       *Salary  `json:"salary"`
	AlternateURL string   `json:"alternate_url"`
	Snippet      Snippet  `json:"snippet"`
	Employment   string   `json:"employment"`
	Schedule     string   `json:"schedule"`
}

type Employer struct {
	Name string `json:"name"`
}

type Salary struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	Currency string `json:"currency"`
}

type Snippet struct {
	PreviewText string `json:"preview_text"`
}

type HHVacanciesResponse struct {
	Items []HHVacancy `json:"items"`
}

type Vacancy struct {
	Title      string
	Company    string
	Salary     string
	URL        string
	Excerpt    string
	Employment string
	Schedule   string
}

type RateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	interval    time.Duration
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		interval: time.Duration(float64(time.Second) / requestsPerSecond),
	}
}

func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRequest)
	if elapsed < r.interval {
		time.Sleep(r.interval - elapsed)
	}
	r.lastRequest = time.Now()
}

type VacancyCache struct {
	mu    sync.Mutex
	items map[string]cacheEntry
	ttl   time.Duration
}

type cacheEntry struct {
	vacancies []Vacancy
	expires   time.Time
}

func NewVacancyCache(ttl time.Duration) *VacancyCache {
	return &VacancyCache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (c *VacancyCache) Get(key string) ([]Vacancy, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok && time.Now().Before(entry.expires) {
		return entry.vacancies, true
	}
	return nil, false
}

func (c *VacancyCache) Set(key string, vacancies []Vacancy) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheEntry{
		vacancies: vacancies,
		expires:   time.Now().Add(c.ttl),
	}
}

func NewFetcher(baseURL string, cacheTTL time.Duration) *Fetcher {
	return &Fetcher{
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(2),
		cache:       NewVacancyCache(cacheTTL),
	}
}

func (f *Fetcher) Search(query string, limit int, employmentTypes, workFormats []string, excludeWords []string) ([]Vacancy, error) {
	cacheKey := fmt.Sprintf("%s:%d:%v:%v", query, limit, employmentTypes, workFormats)
	if cached, ok := f.cache.Get(cacheKey); ok {
		logger.Info().Str("query", query).Int("count", len(cached)).Msg("Using cached vacancies")
		return f.filterVacancies(cached, excludeWords), nil
	}

	f.rateLimiter.Wait()

	params := url.Values{}
	params.Set("text", query)
	params.Set("per_page", fmt.Sprintf("%d", limit))

	for _, et := range employmentTypes {
		params.Add("employment", et)
	}
	for _, wf := range workFormats {
		params.Add("schedule", wf)
	}

	req, err := http.NewRequest("GET", f.baseURL+"/vacancies?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "AI-Service/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vacancies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hh.ru API error: %s", string(body))
	}

	var hhResp HHVacanciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&hhResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	vacancies := make([]Vacancy, 0, len(hhResp.Items))
	for _, v := range hhResp.Items {
		salary := ""
		if v.Salary != nil {
			salary = formatSalary(v.Salary)
		}

		vacancies = append(vacancies, Vacancy{
			Title:      v.Name,
			Company:    v.Employer.Name,
			Salary:     salary,
			URL:        v.AlternateURL,
			Excerpt:    v.Snippet.PreviewText,
			Employment: v.Employment,
			Schedule:   v.Schedule,
		})
	}

	f.cache.Set(cacheKey, vacancies)
	logger.Info().Str("query", query).Int("count", len(vacancies)).Msg("Fetched vacancies from hh.ru")

	filtered := f.filterVacancies(vacancies, excludeWords)
	logger.Info().Int("filtered", len(filtered)).Msg("After exclude_words filtering")

	return filtered, nil
}

func formatSalary(s *Salary) string {
	if s == nil {
		return ""
	}

	from := ""
	to := ""

	if s.From > 0 {
		from = fmt.Sprintf("от %d", s.From)
	}
	if s.To > 0 {
		to = fmt.Sprintf("до %d", s.To)
	}

	if from != "" && to != "" {
		return fmt.Sprintf("%s %s %s", from, to, s.Currency)
	}
	if from != "" {
		return fmt.Sprintf("%s %s", from, s.Currency)
	}
	if to != "" {
		return fmt.Sprintf("%s %s", to, s.Currency)
	}
	return ""
}

func (f *Fetcher) filterVacancies(vacancies []Vacancy, excludeWords []string) []Vacancy {
	if len(excludeWords) == 0 {
		return vacancies
	}

	filtered := make([]Vacancy, 0)
	for _, v := range vacancies {
		text := strings.ToLower(v.Title + " " + v.Excerpt)

		excluded := false
		for _, word := range excludeWords {
			if strings.Contains(text, strings.ToLower(word)) {
				excluded = true
				logger.Debug().Str("word", word).Str("vacancy", v.Title).Msg("Excluded by word")
				break
			}
		}

		if !excluded {
			filtered = append(filtered, v)
		}
	}

	return filtered
}
