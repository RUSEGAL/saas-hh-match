package dbresumes

import (
	"encoding/json"
	"go-api/internal/config/db"
	types_internal "go-api/internal/types/int"
)

func AddResumeDb(userID int64, title, content string) (int64, error) {
	var id int64
	err := db.DB.QueryRow(`
		INSERT INTO resumes (user_id, title, content)
		VALUES ($1, $2, $3) RETURNING id
	`, userID, title, content).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}
func UpdateResumeDb(id, userID int64, title, content string) error {
	_, err := db.DB.Exec(`
		UPDATE resumes
		SET title = $1, content = $2
		WHERE id = $3 AND user_id = $4
	`, title, content, id, userID)

	return err
}
func GetResumesByUserDb(userID int64) ([]types_internal.Resume, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, title, content, created_at
		FROM resumes
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []types_internal.Resume

	for rows.Next() {
		var r types_internal.Resume

		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.Title,
			&r.Content,
			&r.CreatedAt,
		); err != nil {
			return nil, err
		}

		resumes = append(resumes, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resumes, nil
}

func UpdateResumeContent(id int64, title, content string, tags []string, score float64) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(`
		UPDATE resumes
		SET title = $1, content = $2, tags = $3, score = $4
		WHERE id = $5
	`, title, content, string(tagsJSON), score, id)

	return err
}

func GetResumeByID(id int64) (*types_internal.Resume, error) {
	var r types_internal.Resume
	err := db.DB.QueryRow(`
		SELECT id, user_id, title, content, tags, score, created_at
		FROM resumes WHERE id = $1
	`, id).Scan(&r.ID, &r.UserID, &r.Title, &r.Content, &r.Tags, &r.Score, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func DeleteResume(id, userID int64) error {
	_, err := db.DB.Exec(`DELETE FROM resumes WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
