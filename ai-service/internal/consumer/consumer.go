package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"ai-service/internal/ai"
	"ai-service/internal/logger"
	"ai-service/internal/vacancy"
	"ai-service/internal/webhook"
)

type ResumeAnalysisJob struct {
	ResumeID int64  `json:"resume_id"`
	UserID   int64  `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
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

type VacancyMatchPayload struct {
	MatchID int64           `json:"match_id"`
	Job     VacancyMatchJob `json:"job"`
}

type Resume struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type Consumer struct {
	nc             *nats.Conn
	analyzer       *ai.Analyzer
	sender         *webhook.Sender
	resumeWorkers  int
	vacancyWorkers int
	resumeJobs     chan ResumeAnalysisJob
	vacancyJobs    chan VacancyMatchJob
	vacancyMatchID int64
	resumeResults  chan resumeResult
	vacancyResults chan vacancyResult
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc

	vacancyFetcher *vacancy.Fetcher
	vacancyMatcher *vacancy.Matcher
	mainAPIURL     string
	matchesSender  *webhook.MatchesSender
}

type resumeResult struct {
	job    ResumeAnalysisJob
	result *ai.AnalysisResult
	err    error
}

type vacancyResult struct {
	matchID int64
	job     VacancyMatchJob
	matches []vacancy.MatchResult
	err     error
}

func NewConsumer(nc *nats.Conn, analyzer *ai.Analyzer, sender *webhook.Sender, resumeWorkers, vacancyWorkers int) *Consumer {
	if resumeWorkers <= 0 {
		resumeWorkers = runtime.NumCPU()
	}
	if vacancyWorkers <= 0 {
		vacancyWorkers = 2
	}
	return &Consumer{
		nc:             nc,
		analyzer:       analyzer,
		sender:         sender,
		resumeWorkers:  resumeWorkers,
		vacancyWorkers: vacancyWorkers,
		resumeJobs:     make(chan ResumeAnalysisJob, resumeWorkers*10),
		vacancyJobs:    make(chan VacancyMatchJob, vacancyWorkers*10),
		resumeResults:  make(chan resumeResult, resumeWorkers*10),
		vacancyResults: make(chan vacancyResult, vacancyWorkers*10),
	}
}

func (c *Consumer) WithVacancyService(fetcher *vacancy.Fetcher, matcher *vacancy.Matcher, mainAPIURL string, matchesSender *webhook.MatchesSender) *Consumer {
	c.vacancyFetcher = fetcher
	c.vacancyMatcher = matcher
	c.mainAPIURL = mainAPIURL
	c.matchesSender = matchesSender
	return c
}

func (c *Consumer) Start() error {
	js, err := c.nc.JetStream()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get JetStream context")
		return err
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	for i := 0; i < c.resumeWorkers; i++ {
		c.wg.Add(1)
		go c.resumeWorker(i)
	}

	for i := 0; i < c.vacancyWorkers; i++ {
		c.wg.Add(1)
		go c.vacancyWorker(i)
	}

	go c.resumeResultHandler()
	if c.vacancyMatcher != nil {
		go c.vacancyResultHandler()
	}

	_, err = js.Subscribe("resume.analyze", c.handleResumeMessage, nats.Durable("ai-resume-analyzer"))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to subscribe to resume.analyze")
		return err
	}

	logger.Info().Int("resume_workers", c.resumeWorkers).Int("vacancy_workers", c.vacancyWorkers).Msg("Subscribed to resume.analyze")

	if c.vacancyFetcher != nil {
		_, err = js.Subscribe("vacancy.match", c.handleVacancyMessage, nats.Durable("ai-vacancy-matcher"))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to subscribe to vacancy.match")
			return err
		}
		logger.Info().Msg("Subscribed to vacancy.match")
	}

	return nil
}

func (c *Consumer) Stop() {
	c.cancel()
	close(c.resumeJobs)
	if c.vacancyFetcher != nil {
		close(c.vacancyJobs)
	}
	c.wg.Wait()
	close(c.resumeResults)
	if c.vacancyFetcher != nil {
		close(c.vacancyResults)
	}
}

func (c *Consumer) handleResumeMessage(msg *nats.Msg) {
	var job ResumeAnalysisJob
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		logger.Error().Err(err).Str("data", string(msg.Data)).Msg("Failed to parse resume job, acknowledging")
		msg.Ack()
		return
	}

	select {
	case c.resumeJobs <- job:
		msg.Ack()
	case <-time.After(5 * time.Second):
		logger.Warn().Int64("resume_id", job.ResumeID).Msg("Resume job queue full, rejecting")
		msg.Nak()
	}
}

func (c *Consumer) handleVacancyMessage(msg *nats.Msg) {
	var payload VacancyMatchPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		logger.Error().Err(err).Str("data", string(msg.Data)).Msg("Failed to parse vacancy job, acknowledging")
		msg.Ack()
		return
	}

	job := payload.Job
	if job.Limit == 0 {
		job.Limit = 20
	}

	c.vacancyMatchID = payload.MatchID

	select {
	case c.vacancyJobs <- job:
		msg.Ack()
	case <-time.After(5 * time.Second):
		logger.Warn().Int64("resume_id", job.ResumeID).Msg("Vacancy job queue full, rejecting")
		msg.Nak()
	}
}

func (c *Consumer) resumeWorker(id int) {
	defer c.wg.Done()
	logger.Debug().Int("worker", id).Msg("Resume worker started")

	for {
		select {
		case <-c.ctx.Done():
			logger.Debug().Int("worker", id).Msg("Resume worker stopping")
			return
		case job, ok := <-c.resumeJobs:
			if !ok {
				return
			}
			logger.Info().Int64("resume_id", job.ResumeID).Str("title", job.Title).Int("worker", id).Msg("Processing resume job")

			result, err := c.analyzer.Analyze(job.Title, job.Content)
			c.resumeResults <- resumeResult{job: job, result: result, err: err}
		}
	}
}

func (c *Consumer) vacancyWorker(id int) {
	defer c.wg.Done()
	logger.Debug().Int("worker", id).Msg("Vacancy worker started")

	for {
		select {
		case <-c.ctx.Done():
			logger.Debug().Int("worker", id).Msg("Vacancy worker stopping")
			return
		case job, ok := <-c.vacancyJobs:
			if !ok {
				return
			}
			logger.Info().Int64("resume_id", job.ResumeID).Str("query", job.Query).Int("worker", id).Msg("Processing vacancy job")

			matches, err := c.processVacancyMatch(job)
			matchID := c.vacancyMatchID
			c.vacancyResults <- vacancyResult{matchID: matchID, job: job, matches: matches, err: err}
		}
	}
}

func (c *Consumer) processVacancyMatch(job VacancyMatchJob) ([]vacancy.MatchResult, error) {
	resume, err := c.fetchResume(job.ResumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resume: %w", err)
	}

	vacancies, err := c.vacancyFetcher.Search(job.Query, job.Limit, job.EmploymentTypes, job.WorkFormats, job.ExcludeWords)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vacancies: %w", err)
	}

	cacheKey := vacancy.BuildMatchCacheKey(job.ResumeID, job.Query, job.EmploymentTypes, job.WorkFormats, job.ExcludeWords)
	matches, err := c.vacancyMatcher.MatchVacanciesCached(resume.Content, vacancies, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("failed to match vacancies: %w", err)
	}

	return matches, nil
}

func (c *Consumer) fetchResume(resumeID int64) (*Resume, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/resumes/%d", c.mainAPIURL, resumeID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("main API error: %s", string(body))
	}

	var resume Resume
	if err := json.NewDecoder(resp.Body).Decode(&resume); err != nil {
		return nil, fmt.Errorf("failed to parse resume: %w", err)
	}

	return &resume, nil
}

func (c *Consumer) resumeResultHandler() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case res, ok := <-c.resumeResults:
			if !ok {
				return
			}

			if res.err != nil {
				logger.Error().Err(res.err).Int64("resume_id", res.job.ResumeID).Msg("AI analysis failed")
				continue
			}

			logger.Info().Int64("resume_id", res.job.ResumeID).Float64("score", res.result.Score).Msg("Resume analysis complete")

			err := c.sender.Send(res.job.ResumeID, res.job.UserID, res.result)
			if err != nil {
				logger.Error().Err(err).Int64("resume_id", res.job.ResumeID).Msg("Webhook failed")
			}
		}
	}
}

func (c *Consumer) vacancyResultHandler() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case res, ok := <-c.vacancyResults:
			if !ok {
				return
			}

			if res.err != nil {
				logger.Error().Err(res.err).Int64("resume_id", res.job.ResumeID).Msg("Vacancy matching failed")
				continue
			}

			logger.Info().Int64("resume_id", res.job.ResumeID).Int("matches", len(res.matches)).Msg("Vacancy matching complete")

			err := c.matchesSender.Send(res.matchID, res.job.UserID, res.job.ResumeID, res.job.Query, res.matches)
			if err != nil {
				logger.Error().Err(err).Int64("resume_id", res.job.ResumeID).Msg("Matches webhook failed")
			}
		}
	}
}

func ReconnectWithBackoff(url string, maxAttempts int, maxDelay time.Duration) (*nats.Conn, error) {
	var conn *nats.Conn
	var err error

	attempt := 0
	delay := time.Second

	for attempt < maxAttempts {
		conn, err = nats.Connect(url, nats.ReconnectWait(delay))
		if err == nil {
			return conn, nil
		}

		attempt++
		if attempt >= maxAttempts {
			break
		}

		if delay < maxDelay {
			delay *= 2
		}

		logger.Warn().Int("attempt", attempt).Dur("delay", delay).Msg("NATS reconnect failed, retrying")
		time.Sleep(delay)
	}

	return nil, err
}
