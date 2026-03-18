package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"telegram-bot/internal/api"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/internal/logger"
)

type Scheduler struct {
	cron      *cron.Cron
	config    *config.Config
	db        *database.DB
	apiClient *api.APIClient
	jobs      map[int64]cron.EntryID
}

func NewScheduler(cfg *config.Config, apiClient *api.APIClient) *Scheduler {
	s := &Scheduler{
		cron:      cron.New(),
		config:    cfg,
		apiClient: apiClient,
		jobs:      make(map[int64]cron.EntryID),
	}

	if cfg.SchedulerEnabled {
		s.defineJobs()
	}

	return s
}

func (s *Scheduler) defineJobs() {
	s.cron.AddFunc("@every 1m", func() {
		s.processScheduledSearches()
	})
}

func (s *Scheduler) processScheduledSearches() {
	if s.db == nil || s.apiClient == nil {
		return
	}

	schedules, err := s.db.GetAllActiveSchedules()
	if err != nil {
		logger.Error().Err(err).Msg("error getting schedules")
		return
	}

	for _, schedule := range schedules {
		if s.shouldRun(&schedule) {
			go s.executeSearch(&schedule)
		}
	}
}

func (s *Scheduler) shouldRun(schedule *database.UserSchedule) bool {
	if !schedule.IsActive {
		return false
	}

	now := time.Now()
	currentHour := now.Hour()
	currentMin := now.Minute()

	if schedule.Time != "" {
		var hour, min int
		_, err := fmt.Sscanf(schedule.Time, "%d:%d", &hour, &min)
		if err != nil {
			return false
		}

		if currentHour == hour && currentMin == min {
			return true
		}
	}

	return false
}

func (s *Scheduler) executeSearch(schedule *database.UserSchedule) {
	logger.Info().Int64("user_id", schedule.UserID).Str("query", schedule.Query).Msg("executing scheduled search")

	job, err := s.apiClient.MatchVacancies(schedule.UserID, &api.MatchRequest{
		ResumeID: schedule.ResumeID,
		Query:    schedule.Query,
		Filters:  nil,
	})
	if err != nil {
		logger.Error().Err(err).Int64("user_id", schedule.UserID).Msg("failed to start vacancy search")
		return
	}

	logger.Info().Int64("user_id", schedule.UserID).Int64("job_id", job.ID).Msg("started scheduled vacancy search")
	s.db.UpdateLastRun(schedule.UserID)
}

func (s *Scheduler) AddJob(userID int64, cronExpr string) error {
	job := cron.FuncJob(func() {
		logger.Info().Int64("user_id", userID).Msg("running scheduled job")
	})

	entryID, err := s.cron.AddJob(cronExpr, job)
	if err != nil {
		return fmt.Errorf("failed to add job: %w", err)
	}

	s.jobs[userID] = entryID
	return nil
}

func (s *Scheduler) RemoveJob(userID int64) {
	if entryID, ok := s.jobs[userID]; ok {
		s.cron.Remove(entryID)
		delete(s.jobs, userID)
	}
}

func (s *Scheduler) Start() {
	logger.Info().Msg("starting scheduler")
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	logger.Info().Msg("stopping scheduler")
	s.cron.Stop()
}

func parseWeekday(day string) time.Weekday {
	switch strings.ToLower(day) {
	case "mon", "monday":
		return time.Monday
	case "tue", "tuesday":
		return time.Tuesday
	case "wed", "wednesday":
		return time.Wednesday
	case "thu", "thursday":
		return time.Thursday
	case "fri", "friday":
		return time.Friday
	case "sat", "saturday":
		return time.Saturday
	case "sun", "sunday":
		return time.Sunday
	default:
		return -1
	}
}
