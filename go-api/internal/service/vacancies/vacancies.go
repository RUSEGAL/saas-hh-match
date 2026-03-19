package vacancies

import (
	"errors"

	"go-api/internal/cache"
	"go-api/internal/nats"
	dbvacancies "go-api/internal/repository/vacancies"
	types_internal "go-api/internal/types/int"
)

func CreateMatchJob(userID, resumeID int64, query string, limit int, excludeWords, employmentTypes, workFormats []string) (*types_internal.VacancyMatch, error) {
	if query == "" {
		return nil, errors.New("query is required")
	}
	if resumeID <= 0 {
		return nil, errors.New("resume_id is required")
	}

	if limit <= 0 {
		limit = 20
	}

	id, err := dbvacancies.CreateVacancyMatch(userID, resumeID, query)
	if err != nil {
		return nil, err
	}

	job := types_internal.VacancyMatchJob{
		ResumeID:        resumeID,
		UserID:          userID,
		Query:           query,
		Limit:           limit,
		ExcludeWords:    excludeWords,
		EmploymentTypes: employmentTypes,
		WorkFormats:     workFormats,
	}
	nats.PublishVacancyMatch(id, job)

	return &types_internal.VacancyMatch{
		ID:       id,
		UserID:   userID,
		ResumeID: resumeID,
		Query:    query,
		Status:   "pending",
	}, nil
}

func GetMatchWithResults(matchID int64) (*types_internal.VacancyMatch, []types_internal.VacancyMatchResult, error) {
	match, err := dbvacancies.GetVacancyMatchByID(matchID)
	if err != nil {
		return nil, nil, err
	}

	results, err := dbvacancies.GetVacancyMatchResults(matchID)
	if err != nil {
		return nil, nil, err
	}

	return match, results, nil
}

func GetUserMatches(userID int64) ([]types_internal.VacancyMatch, error) {
	return dbvacancies.GetUserMatches(userID)
}

func ProcessWebhookResult(result *types_internal.VacancyWebhookResult) error {
	if err := dbvacancies.UpdateVacancyMatchStatus(result.MatchID, "completed"); err != nil {
		return err
	}

	if len(result.Matches) > 0 {
		if err := dbvacancies.SaveVacancyMatchResults(result.MatchID, result.Matches); err != nil {
			return err
		}
	}

	cache.Delete("vacancy_matches:all")
	cache.Delete("vacancy_matches:user:" + string(rune(result.UserID)))

	return nil
}

func SaveResponse(userID, vacancyID int64) error {
	return dbvacancies.SaveVacancyResponse(userID, vacancyID)
}

func SaveView(userID, vacancyID int64) error {
	return dbvacancies.SaveVacancyView(userID, vacancyID)
}
