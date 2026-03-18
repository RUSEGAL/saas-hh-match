package database

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

type BotUser struct {
	ID         int64  `db:"id"`
	TelegramID int64  `db:"telegram_id"`
	Username   string `db:"username"`
	CreatedAt  string `db:"created_at"`
	IsActive   bool   `db:"is_active"`
}

type UserSchedule struct {
	ID       int64  `db:"id"`
	UserID   int64  `db:"user_id"`
	CronExpr string `db:"cron_expr"`
	Time     string `db:"time"`
	Days     string `db:"days"`
	ResumeID int64  `db:"resume_id"`
	Query    string `db:"query"`
	Filters  string `db:"filters"`
	IsActive bool   `db:"is_active"`
	LastRun  string `db:"last_run"`
}

func NewDB(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL not provided")
	}

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db}, nil
}

func createTables(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS bot_users (
		id SERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		username TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT true
	);

	CREATE TABLE IF NOT EXISTS user_schedules (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		cron_expr TEXT,
		time TEXT,
		days TEXT,
		resume_id BIGINT,
		query TEXT,
		filters TEXT,
		is_active BOOLEAN DEFAULT true,
		last_run TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES bot_users(telegram_id)
	);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) CreateUser(telegramID int64, username string) error {
	query := `
	INSERT INTO bot_users (telegram_id, username)
	VALUES ($1, $2)
	ON CONFLICT (telegram_id) DO NOTHING
	`
	_, err := db.Exec(query, telegramID, username)
	return err
}

func (db *DB) GetUser(telegramID int64) (*BotUser, error) {
	var user BotUser
	query := `SELECT * FROM bot_users WHERE telegram_id = $1`
	err := db.Get(&user, query, telegramID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) SaveSchedule(userID int64, cronExpr, time string, resumeID int64, query string) error {
	queryStmt := `
	INSERT INTO user_schedules (user_id, cron_expr, resume_id, query, is_active)
	VALUES ($1, $2, $3, $4, true)
	ON CONFLICT (user_id) DO UPDATE SET
		cron_expr = EXCLUDED.cron_expr,
		resume_id = EXCLUDED.resume_id,
		query = EXCLUDED.query,
		is_active = true
	`
	_, err := db.Exec(queryStmt, userID, cronExpr, resumeID, query)
	return err
}

func (db *DB) GetSchedule(userID int64) (*UserSchedule, error) {
	var schedule UserSchedule
	query := `SELECT * FROM user_schedules WHERE user_id = $1 AND is_active = true`
	err := db.Get(&schedule, query, userID)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (db *DB) GetAllActiveSchedules() ([]UserSchedule, error) {
	var schedules []UserSchedule
	query := `SELECT * FROM user_schedules WHERE is_active = true`
	err := db.Select(&schedules, query)
	return schedules, err
}

func (db *DB) DeleteSchedule(userID int64) error {
	query := `UPDATE user_schedules SET is_active = false WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

func (db *DB) UpdateLastRun(userID int64) error {
	query := `UPDATE user_schedules SET last_run = CURRENT_TIMESTAMP WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	return err
}

func (db *DB) SaveUserFilters(userID int64, filters interface{}) error {
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO user_schedules (user_id, filters)
	VALUES ($1, $2)
	ON CONFLICT (user_id) DO UPDATE SET filters = EXCLUDED.filters
	`
	_, err = db.Exec(query, userID, filtersJSON)
	return err
}

func (db *DB) GetUserFilters(userID int64) (map[string]interface{}, error) {
	var filters string
	query := `SELECT filters FROM user_schedules WHERE user_id = $1`
	err := db.Get(&filters, query, userID)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(filters), &result)
	return result, err
}
