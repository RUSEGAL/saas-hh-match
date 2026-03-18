package dbadmin

import (
	"go-api/internal/config/db"
	types_internal "go-api/internal/types/int"
)

func GetAllUsers() ([]types_internal.User, error) {
	rows, err := db.DB.Query("SELECT id, username FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []types_internal.User
	for rows.Next() {
		var u types_internal.User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func GetAllStats() ([]types_internal.UserStats, error) {
	rows, err := db.DB.Query(`
		SELECT 
			u.id, u.username,
			COALESCE(r.count, 0),
			COALESCE(p.count, 0),
			COALESCE(p.total_amount, 0)
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as count FROM resumes GROUP BY user_id
		) r ON u.id = r.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as count, SUM(amount) as total_amount 
			FROM payments GROUP BY user_id
		) p ON u.id = p.user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []types_internal.UserStats
	for rows.Next() {
		var s types_internal.UserStats
		if err := rows.Scan(&s.ID, &s.Username, &s.Resumes, &s.Payments, &s.TotalAmount); err != nil {
			continue
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func GetUserStats(userID int64) (*types_internal.UserStats, error) {
	var s types_internal.UserStats
	err := db.DB.QueryRow(`
		SELECT 
			u.id, u.username,
			COALESCE((SELECT COUNT(*) FROM resumes WHERE user_id = u.id), 0),
			COALESCE((SELECT COUNT(*) FROM payments WHERE user_id = u.id), 0),
			COALESCE((SELECT SUM(amount) FROM payments WHERE user_id = u.id), 0)
		FROM users u WHERE u.id = $1
	`, userID).Scan(&s.ID, &s.Username, &s.Resumes, &s.Payments, &s.TotalAmount)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetUserResumes(userID int64) ([]types_internal.ResumeWithUser, error) {
	rows, err := db.DB.Query(`
		SELECT r.id, u.username, r.title, r.content, r.created_at
		FROM resumes r
		JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []types_internal.ResumeWithUser
	for rows.Next() {
		var r types_internal.ResumeWithUser
		if err := rows.Scan(&r.ID, &r.Username, &r.Title, &r.Content, &r.CreatedAt); err != nil {
			continue
		}
		resumes = append(resumes, r)
	}
	return resumes, rows.Err()
}

func GetUserPayments(userID int64) ([]types_internal.PaymentWithUser, error) {
	rows, err := db.DB.Query(`
		SELECT p.id, u.username, p.amount, p.status, p.provider, p.created_at
		FROM payments p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []types_internal.PaymentWithUser
	for rows.Next() {
		var p types_internal.PaymentWithUser
		if err := rows.Scan(&p.ID, &p.Username, &p.Amount, &p.Status, &p.Provider, &p.CreatedAt); err != nil {
			continue
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
