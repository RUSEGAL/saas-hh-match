package types_internal

import "time"

type Payment struct {
	ID        int64
	UserID    int64
	Amount    int64
	Status    string
	Provider  string
	CreatedAt time.Time
}
type Resume struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      string    `json:"tags"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
type UserStats struct {
	ID          int64  `json:"user_id"`
	Username    string `json:"username"`
	Resumes     int    `json:"resumes"`
	Payments    int    `json:"payments"`
	TotalAmount int    `json:"total_amount"`
}
type ResumeWithUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
type PaymentWithUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

type VacancyMatch struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ResumeID  int64     `json:"resume_id"`
	Query     string    `json:"query"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type VacancyMatchResult struct {
	ID           int64   `json:"id"`
	MatchID      int64   `json:"match_id"`
	VacancyTitle string  `json:"vacancy_title"`
	Company      string  `json:"company"`
	Score        float64 `json:"score"`
	URL          string  `json:"url"`
	Salary       string  `json:"salary"`
	Excerpt      string  `json:"excerpt"`
}

type VacancyMatchJob struct {
	ResumeID        int64    `json:"resume_id"`
	UserID          int64    `json:"user_id"`
	Query           string   `json:"query"`
	Limit           int      `json:"limit"`
	ExcludeWords    []string `json:"exclude_words"`
	EmploymentTypes []string `json:"employment_types"`
	WorkFormats     []string `json:"work_formats"`
}

type VacancyWebhookResult struct {
	UserID   int64                     `json:"user_id"`
	ResumeID int64                     `json:"resume_id"`
	MatchID  int64                     `json:"match_id"`
	Query    string                    `json:"query"`
	Matches  []VacancyMatchResultInput `json:"matches"`
}

type VacancyMatchResultInput struct {
	VacancyTitle string  `json:"vacancy_title"`
	Company      string  `json:"vacancy_company"`
	Score        float64 `json:"score"`
	URL          string  `json:"url"`
	Salary       string  `json:"salary"`
	Excerpt      string  `json:"excerpt"`
}
