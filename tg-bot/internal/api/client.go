package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"telegram-bot/internal/config"
	"time"
)

type APIClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type Payment struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	Amount    int       `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Resume struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AIScore   float64   `json:"ai_score,omitempty"`
	Analyzed  bool      `json:"analyzed"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Vacancy struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Company    string  `json:"company"`
	MatchScore float64 `json:"match_score"`
	Salary     string  `json:"salary"`
	Location   string  `json:"location"`
	URL        string  `json:"url"`
	Employment string  `json:"employment,omitempty"`
	Format     string  `json:"format,omitempty"`
}

type MatchRequest struct {
	ResumeID int64           `json:"resume_id"`
	Query    string          `json:"query"`
	Filters  *VacancyFilters `json:"filters,omitempty"`
}

type VacancyFilters struct {
	Employment []string `json:"employment,omitempty"`
	Format     []string `json:"format,omitempty"`
	Exclude    []string `json:"exclude,omitempty"`
}

type MatchJob struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type MatchResult struct {
	JobID     int64     `json:"job_id"`
	Vacancies []Vacancy `json:"vacancies"`
	Total     int       `json:"total"`
}

type UserStats struct {
	ResumesCount    int `json:"resumes_count"`
	SearchesCount   int `json:"searches_count"`
	VacanciesFound  int `json:"vacancies_found"`
	VacanciesViewed int `json:"vacancies_viewed"`
	ResponsesCount  int `json:"responses_count"`
	MonthResumes    int `json:"month_resumes"`
	MonthSearches   int `json:"month_searches"`
	MonthVacancies  int `json:"month_vacancies"`
}

type CreateResumeResponse struct {
	Resume *Resume `json:"resume"`
}

type PaymentLinkRequest struct {
	Duration int `json:"duration"`
}

type PaymentLinkResponse struct {
	URL string `json:"url"`
}

func NewAPIClient(cfg *config.Config) *APIClient {
	return &APIClient{
		BaseURL: cfg.APIURL,
		Token:   cfg.APIToken,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var result bytes.Buffer
	if _, err := result.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func (c *APIClient) GetUserPayments(userID int64) ([]Payment, error) {
	type PaymentsResponse struct {
		Payments []Payment `json:"payments"`
	}

	data, err := c.doRequest("GET", "/api/payments/me", nil)
	if err != nil {
		return nil, err
	}
	resp := &PaymentsResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp.Payments, nil
}

func (c *APIClient) GetPaymentStatus(userID int64) (*Payment, error) {
	type PaymentResponse struct {
		IsActive      bool   `json:"is_active"`
		ExpiresAt     string `json:"expires_at,omitempty"`
		DaysRemaining int    `json:"days_remaining"`
	}

	data, err := c.doRequest("GET", "/api/user/payment-status", nil)
	if err != nil {
		return nil, err
	}
	resp := &PaymentResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return &Payment{
		ID:        0,
		Status:    "active",
		Amount:    0,
		CreatedAt: time.Time{},
		ExpiresAt: time.Time{},
	}, nil
}

func (c *APIClient) GetCurrentPayment(userID int64) (*Payment, error) {
	return c.GetPaymentStatus(userID)
}

func (c *APIClient) IsSubscribed(userID int64) (bool, *Payment, error) {
	payment, err := c.GetPaymentStatus(userID)
	if err != nil {
		return false, nil, err
	}
	if payment == nil {
		return false, nil, nil
	}
	return payment.Status == "active", payment, nil
}

func (c *APIClient) CreateResume(userID int64, title, content string) (*Resume, error) {
	type CreateResumeReq struct {
		UserID   int64  `json:"user_id"`
		Title    string `json:"title"`
		Content  string `json:"content"`
		Telegram bool   `json:"telegram"`
	}

	data, err := c.doRequest("POST", "/api/resumes", CreateResumeReq{
		UserID:   userID,
		Title:    title,
		Content:  content,
		Telegram: true,
	})
	if err != nil {
		return nil, err
	}
	resp := &CreateResumeResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp.Resume, nil
}

func (c *APIClient) GetResumes(userID int64) ([]Resume, error) {
	type ResumesResponse struct {
		Resumes []Resume `json:"resumes"`
	}

	data, err := c.doRequest("GET", fmt.Sprintf("/api/resumes?user_id=%d", userID), nil)
	if err != nil {
		return nil, err
	}
	resp := &ResumesResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp.Resumes, nil
}

func (c *APIClient) GetResume(resumeID int64) (*Resume, error) {
	type ResumeResponse struct {
		Resume Resume `json:"resume"`
	}

	data, err := c.doRequest("GET", fmt.Sprintf("/api/resumes/%d", resumeID), nil)
	if err != nil {
		return nil, err
	}
	resp := &ResumeResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return &resp.Resume, nil
}

func (c *APIClient) UpdateResume(resumeID int64, title, content string) (*Resume, error) {
	type UpdateReq struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	data, err := c.doRequest("PATCH", fmt.Sprintf("/api/resumes/%d", resumeID), UpdateReq{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, err
	}
	resp := &CreateResumeResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp.Resume, nil
}

func (c *APIClient) DeleteResume(resumeID int64) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/resumes/%d", resumeID), nil)
	return err
}

func (c *APIClient) MatchVacancies(userID int64, req *MatchRequest) (*MatchJob, error) {
	type MatchReq struct {
		UserID   int64           `json:"user_id"`
		ResumeID int64           `json:"resume_id"`
		Query    string          `json:"query"`
		Filters  *VacancyFilters `json:"filters,omitempty"`
	}

	data, err := c.doRequest("POST", "/api/vacancies/match", MatchReq{
		UserID:   userID,
		ResumeID: req.ResumeID,
		Query:    req.Query,
		Filters:  req.Filters,
	})
	if err != nil {
		return nil, err
	}
	resp := &struct {
		Job *MatchJob `json:"job"`
	}{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp.Job, nil
}

func (c *APIClient) GetMatchResults(matchID int64) (*MatchResult, error) {
	type ResultResponse struct {
		Match   MatchJob        `json:"match"`
		Results []VacancyResult `json:"results"`
	}

	data, err := c.doRequest("GET", fmt.Sprintf("/api/vacancies/matches/%d", matchID), nil)
	if err != nil {
		return nil, err
	}
	resp := &ResultResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return &MatchResult{
		JobID:     matchID,
		Vacancies: convertToVacancies(resp.Results),
		Total:     len(resp.Results),
	}, nil
}

type VacancyResult struct {
	VacancyTitle string  `json:"vacancy_title"`
	Company      string  `json:"company"`
	Score        float64 `json:"score"`
	URL          string  `json:"url"`
	Salary       string  `json:"salary"`
	Excerpt      string  `json:"excerpt"`
}

func convertToVacancies(results []VacancyResult) []Vacancy {
	vacancies := make([]Vacancy, 0, len(results))
	for _, r := range results {
		vacancies = append(vacancies, Vacancy{
			ID:         0,
			Title:      r.VacancyTitle,
			Company:    r.Company,
			MatchScore: r.Score,
			Salary:     r.Salary,
			Location:   "",
			URL:        r.URL,
		})
	}
	return vacancies
}

func (c *APIClient) GetUserStats(userID int64) (*UserStats, error) {
	type StatsResponse struct {
		ResumesCount   int    `json:"resumes_count"`
		SearchesCount  int    `json:"searches_count"`
		MatchesCount   int    `json:"matches_count"`
		PaymentsCount  int    `json:"payments_count"`
		LastSearchDate string `json:"last_search_date,omitempty"`
	}

	data, err := c.doRequest("GET", "/api/user/stats", nil)
	if err != nil {
		return nil, err
	}
	resp := &StatsResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return &UserStats{
		ResumesCount:    resp.ResumesCount,
		SearchesCount:   resp.SearchesCount,
		VacanciesFound:  resp.MatchesCount,
		VacanciesViewed: 0,
		ResponsesCount:  0,
		MonthResumes:    0,
		MonthSearches:   0,
		MonthVacancies:  0,
	}, nil
}

func (c *APIClient) GetPaymentLink(userID int64, duration int) (string, error) {
	type PaymentReq struct {
		Duration int `json:"duration"`
	}

	data, err := c.doRequest("POST", "/api/payments", PaymentReq{
		Duration: duration,
	})
	if err != nil {
		return "", err
	}
	resp := &PaymentLinkResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return "", err
	}
	return resp.URL, nil
}

func (c *APIClient) SaveVacancyResponse(userID, vacancyID int64) error {
	type SaveReq struct {
		UserID    int64 `json:"user_id"`
		VacancyID int64 `json:"vacancy_id"`
	}

	_, err := c.doRequest("POST", "/api/vacancies/response", SaveReq{
		UserID:    userID,
		VacancyID: vacancyID,
	})
	return err
}

func (c *APIClient) SaveVacancyView(userID, vacancyID int64) error {
	type ViewReq struct {
		UserID    int64 `json:"user_id"`
		VacancyID int64 `json:"vacancy_id"`
	}

	_, err := c.doRequest("POST", "/api/vacancies/view", ViewReq{
		UserID:    userID,
		VacancyID: vacancyID,
	})
	return err
}
