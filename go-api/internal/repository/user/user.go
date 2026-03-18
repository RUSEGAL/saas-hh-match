package dbuser

import (
	"time"

	"go-api/internal/config/db"
)

type UserStats struct {
	UserID         int64  `json:"user_id"`
	ResumesCount   int    `json:"resumes_count"`
	SearchesCount  int    `json:"searches_count"`
	MatchesCount   int    `json:"matches_count"`
	PaymentsCount  int    `json:"payments_count"`
	LastSearchDate string `json:"last_search_date,omitempty"`
}

type PaymentStatus struct {
	IsActive      bool   `json:"is_active"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	DaysRemaining int    `json:"days_remaining"`
	LastPaymentAt string `json:"last_payment_at,omitempty"`
}

func GetUserStats(userID int64) (*UserStats, error) {
	stats := &UserStats{UserID: userID}

	err := db.DB.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM resumes WHERE user_id = $1) as resumes_count,
			(SELECT COUNT(*) FROM vacancy_matches WHERE user_id = $1) as searches_count,
			(SELECT COUNT(*) FROM vacancy_match_results mr 
				JOIN vacancy_matches m ON mr.match_id = m.id 
				WHERE m.user_id = $1) as matches_count,
			(SELECT COUNT(*) FROM payments WHERE user_id = $1) as payments_count,
			(SELECT MAX(created_at) FROM vacancy_matches WHERE user_id = $1) as last_search
	`, userID).Scan(
		&stats.ResumesCount,
		&stats.SearchesCount,
		&stats.MatchesCount,
		&stats.PaymentsCount,
		&stats.LastSearchDate,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func GetPaymentStatus(userID int64) (*PaymentStatus, error) {
	status := &PaymentStatus{}

	var lastPaymentTime time.Time
	err := db.DB.QueryRow(`
		SELECT created_at FROM payments 
		WHERE user_id = $1 AND status = 'completed'
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&lastPaymentTime)

	if err != nil {
		status.IsActive = false
		return status, nil
	}

	status.LastPaymentAt = lastPaymentTime.Format("2006-01-02")

	expiresAt := lastPaymentTime.AddDate(0, 0, 30)
	status.ExpiresAt = expiresAt.Format("2006-01-02")

	daysRemaining := int(time.Until(expiresAt).Hours() / 24)
	status.DaysRemaining = daysRemaining

	if daysRemaining > 0 {
		status.IsActive = true
	}

	return status, nil
}
