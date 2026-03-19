-- Database initialization for go-api
-- This runs on first container start

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tokens table
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token);

-- Payments table
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    provider VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);

-- Resumes table
CREATE TABLE IF NOT EXISTS resumes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500),
    content TEXT,
    tags TEXT,
    score REAL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id);

-- Vacancy matches table
CREATE TABLE IF NOT EXISTS vacancy_matches (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resume_id INTEGER REFERENCES resumes(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vacancy_matches_user_id ON vacancy_matches(user_id);

-- Vacancy match results table
CREATE TABLE IF NOT EXISTS vacancy_match_results (
    id SERIAL PRIMARY KEY,
    match_id INTEGER NOT NULL REFERENCES vacancy_matches(id) ON DELETE CASCADE,
    vacancy_title VARCHAR(500),
    vacancy_company VARCHAR(255),
    score REAL,
    url TEXT,
    salary VARCHAR(100),
    excerpt TEXT
);

CREATE INDEX IF NOT EXISTS idx_vacancy_match_results_match_id ON vacancy_match_results(match_id);

-- Admin user (password: admin123 - CHANGE THIS!)
INSERT INTO users (username, is_admin) 
VALUES ('admin', TRUE) 
ON CONFLICT (username) DO NOTHING;

-- Vacancy responses table
CREATE TABLE IF NOT EXISTS vacancy_responses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vacancy_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, vacancy_id)
);

-- Vacancy views table
CREATE TABLE IF NOT EXISTS vacancy_views (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vacancy_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, vacancy_id)
);
