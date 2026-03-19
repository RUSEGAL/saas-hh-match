package resumes

import (
	"go-api/internal/nats"
	dbresumes "go-api/internal/repository/resumes"
	dbuser "go-api/internal/repository/user"
	types_external "go-api/internal/types/external"
	types_internal "go-api/internal/types/int"
)

func CreateResume(userID int64, req *types_external.ResumeRequest) (int64, error) {
	dbuser.EnsureUser(userID, "")

	id, err := dbresumes.AddResumeDb(userID, req.Title, req.Content)
	if err != nil {
		return 0, err
	}

	if nats.Conn != nil {
		nats.PublishResumeAnalysis(nats.ResumeAnalysisJob{
			ResumeID: id,
			UserID:   userID,
			Title:    req.Title,
			Content:  req.Content,
		})
	}

	return id, nil
}

func UpdateResume(id, userID int64, req *types_external.ResumeRequest) error {
	return dbresumes.UpdateResumeDb(id, userID, req.Title, req.Content)
}

func GetResumeByUser(userID int64) ([]types_internal.Resume, error) {
	return dbresumes.GetResumesByUserDb(userID)
}
