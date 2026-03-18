package nats

import (
	"encoding/json"
	"os"

	"github.com/nats-io/nats.go"
	types_internal "go-api/internal/types/int"
)

var Conn *nats.Conn

func Init() error {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	var err error
	Conn, err = nats.Connect(url)
	return err
}

type ResumeAnalysisJob struct {
	ResumeID int64  `json:"resume_id"`
	UserID   int64  `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

func PublishResumeAnalysis(job ResumeAnalysisJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return Conn.Publish("resume.analyze", data)
}

func PublishVacancyMatch(matchID int64, job types_internal.VacancyMatchJob) error {
	payload := map[string]interface{}{
		"match_id": matchID,
		"job":      job,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return Conn.Publish("vacancy.match", data)
}

func Close() {
	if Conn != nil {
		Conn.Close()
	}
}
