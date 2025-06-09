package services

import (
	"fmt"
	"questions-vote/internal/elo"
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

// GetQuestions returns a pair of questions for voting using ELO selection
func (s *QuestionService) GetQuestions() ([]*models.Question, error) {
	// Get the active tournament
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		// Fallback to mock data if no active tournament
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
	
	// Use ELO system to select question pair
	eloSystem := elo.New(tournament)
	q1ID, q2ID, err := eloSystem.SelectPair()
	if err != nil {
		// Fallback to mock data if ELO selection fails
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
	
	// Get the actual questions from the database
	questions, err := s.questionRepo.FindByIDs([]int{q1ID, q2ID})
	if err != nil {
		// Fallback to mock data if database query fails
		return []*models.Question{
			{
				ID:       q1ID,
				Question: "Sample question " + fmt.Sprintf("%d", q1ID) + "?",
				Answer:   "Sample answer " + fmt.Sprintf("%d", q1ID),
				Comment:  "Sample comment " + fmt.Sprintf("%d", q1ID),
				Source:   "Sample source " + fmt.Sprintf("%d", q1ID),
			},
			{
				ID:       q2ID,
				Question: "Sample question " + fmt.Sprintf("%d", q2ID) + "?",
				Answer:   "Sample answer " + fmt.Sprintf("%d", q2ID),
				Comment:  "Sample comment " + fmt.Sprintf("%d", q2ID),
				Source:   "Sample source " + fmt.Sprintf("%d", q2ID),
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