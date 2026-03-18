package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var DBReplica *sql.DB

func Init() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "goapi")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5)) * time.Minute

	DB.SetMaxOpenConns(maxOpenConns)
	DB.SetMaxIdleConns(maxIdleConns)
	DB.SetConnMaxLifetime(connMaxLifetime)

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected")

	replicaHost := os.Getenv("DB_REPLICA_HOST")
	if replicaHost != "" {
		replicaPort := getEnv("DB_REPLICA_PORT", "5432")
		replicaDsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			replicaHost, replicaPort, user, password, dbname,
		)
		DBReplica, err = sql.Open("postgres", replicaDsn)
		if err != nil {
			log.Printf("Warning: failed to connect to replica: %v", err)
		} else {
			DBReplica.SetMaxOpenConns(maxOpenConns)
			DBReplica.SetMaxIdleConns(maxIdleConns)
			DBReplica.SetConnMaxLifetime(connMaxLifetime)
			log.Println("Database replica connected")
		}
	}

	createTables()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			telegram_id BIGINT UNIQUE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			amount INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL,
			provider VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS resumes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(500),
			content TEXT,
			tags TEXT,
			score REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS vacancy_matches (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			resume_id INTEGER NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
			query TEXT NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS vacancy_match_results (
			id SERIAL PRIMARY KEY,
			match_id INTEGER NOT NULL REFERENCES vacancy_matches(id) ON DELETE CASCADE,
			vacancy_title VARCHAR(500),
			vacancy_company VARCHAR(255),
			score REAL,
			url TEXT,
			salary VARCHAR(100),
			excerpt TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token)`,
		`CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vacancy_matches_user_id ON vacancy_matches(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_vacancy_match_results_match_id ON vacancy_match_results(match_id)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			log.Printf("Warning: failed to execute query: %v", err)
		}
	}
}
