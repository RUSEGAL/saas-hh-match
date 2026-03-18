package dbtokens

import (
	"database/sql"
	"errors"
	"go-api/internal/config/db"
	"go-api/internal/helpers"
)

func AddTokensDb(username string, telegramID int64) (string, error) {
	var userID int

	err := db.DB.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		username,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var newID int64
			err = db.DB.QueryRow(
				"INSERT INTO users (username, telegram_id) VALUES ($1, $2) RETURNING id",
				username,
				telegramID,
			).Scan(&newID)
			if err != nil {
				return "", err
			}
			userID = int(newID)
		} else {
			return "", err
		}
	} else if telegramID > 0 {
		db.DB.Exec(
			"UPDATE users SET telegram_id = $1 WHERE id = $2",
			telegramID,
			userID,
		)
	}

	var count int
	err = db.DB.QueryRow(
		"SELECT COUNT(*) FROM tokens WHERE user_id = $1",
		userID,
	).Scan(&count)

	if err != nil {
		return "", err
	}

	if count >= 5 {
		return "", errors.New("token limit reached (max 5)")
	}

	token := helpers.RandStr(64)

	_, err = db.DB.Exec(
		"INSERT INTO tokens (user_id, token) VALUES ($1, $2)",
		userID,
		token,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func FindTokenDb(token string) (int, error) {
	var userID int

	err := db.DB.QueryRow(
		"SELECT user_id FROM tokens WHERE token = $1",
		token,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("token not found")
		}
		return 0, err
	}

	return userID, nil
}
