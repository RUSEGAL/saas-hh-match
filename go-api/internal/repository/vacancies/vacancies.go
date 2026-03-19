package dbvacancies

import (
	"go-api/internal/config/db"
	types_internal "go-api/internal/types/int"
)

func CreateVacancyMatch(userID, resumeID int64, query string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`
		INSERT INTO vacancy_matches (user_id, resume_id, query, status)
		VALUES ($1, $2, $3, 'pending') RETURNING id
	`, userID, resumeID, query).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdateVacancyMatchStatus(id int64, status string) error {
	_, err := db.DB.Exec(`
		UPDATE vacancy_matches SET status = $1 WHERE id = $2
	`, status, id)
	return err
}

func GetVacancyMatchByID(id int64) (*types_internal.VacancyMatch, error) {
	var m types_internal.VacancyMatch
	err := db.DB.QueryRow(`
		SELECT id, user_id, resume_id, query, status, created_at
		FROM vacancy_matches WHERE id = $1
	`, id).Scan(&m.ID, &m.UserID, &m.ResumeID, &m.Query, &m.Status, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetVacancyMatchResults(matchID int64) ([]types_internal.VacancyMatchResult, error) {
	rows, err := db.DB.Query(`
		SELECT id, match_id, vacancy_title, vacancy_company, score, url, salary, excerpt
		FROM vacancy_match_results WHERE match_id = $1
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types_internal.VacancyMatchResult
	for rows.Next() {
		var r types_internal.VacancyMatchResult
		if err := rows.Scan(&r.ID, &r.MatchID, &r.VacancyTitle, &r.Company, &r.Score, &r.URL, &r.Salary, &r.Excerpt); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func SaveVacancyMatchResults(matchID int64, results []types_internal.VacancyMatchResultInput) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, r := range results {
		_, err := tx.Exec(`
			INSERT INTO vacancy_match_results 
			(match_id, vacancy_title, vacancy_company, score, url, salary, excerpt)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, matchID, r.VacancyTitle, r.Company, r.Score, r.URL, r.Salary, r.Excerpt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetUserMatches(userID int64) ([]types_internal.VacancyMatch, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, resume_id, query, status, created_at
		FROM vacancy_matches WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []types_internal.VacancyMatch
	for rows.Next() {
		var m types_internal.VacancyMatch
		if err := rows.Scan(&m.ID, &m.UserID, &m.ResumeID, &m.Query, &m.Status, &m.CreatedAt); err != nil {
			continue
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

func SaveVacancyResponse(userID, vacancyID int64) error {
	_, err := db.DB.Exec(`
		INSERT INTO vacancy_responses (user_id, vacancy_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, vacancyID)
	return err
}

func SaveVacancyView(userID, vacancyID int64) error {
	_, err := db.DB.Exec(`
		INSERT INTO vacancy_views (user_id, vacancy_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, vacancyID)
	return err
}
