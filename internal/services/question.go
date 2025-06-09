package services

import (
	"questions-vote/internal/models"
)

// QuestionService handles question-related operations
type QuestionService struct {
	questionRepo   *models.QuestionRepository
	tournamentRepo *models.TournamentRepository
}

// NewQuestionService creates a new question service
func NewQuestionService() *QuestionService {
	return &QuestionService{
		questionRepo:   models.NewQuestionRepository(),
		tournamentRepo: models.NewTournamentRepository(),
	}
}

// GetQuestions returns a pair of questions for voting
func (s *QuestionService) GetQuestions() ([]*models.Question, error) {
	// TODO: Implement ELO-based question selection
	// For now, get any two questions from the database
	// This is a temporary implementation - should use ELO selection
	
	// Get some questions (mock implementation for now)
	questions, err := s.questionRepo.FindByIDs([]int{1, 2})
	if err != nil {
		// Fallback to mock data if database query fails
		return []*models.Question{
			{
				ID:       1,
				Question: "Sample question 1?",
				Answer:   "Sample answer 1",
				Comment:  "Sample comment 1",
				Source:   "Sample source 1",
			},
			{
				ID:       2,
				Question: "Sample question 2?",
				Answer:   "Sample answer 2",
				Comment:  "Sample comment 2",
				Source:   "Sample source 2",
			},
		}, nil
	}
	
	return questions, nil
}

// FindQuestions finds questions by IDs
func (s *QuestionService) FindQuestions(ids []int) ([]*models.Question, error) {
	return s.questionRepo.FindByIDs(ids)
}

// GetQuestionsCount returns the number of questions in the active tournament
func (s *QuestionService) GetQuestionsCount() (int, error) {
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		return 1000, nil // Fallback value
	}
	return tournament.QuestionsCount, nil
}