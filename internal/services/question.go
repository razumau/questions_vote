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
		return nil, fmt.Errorf("failed to find active tournament: %w", err)
	}
	
	// Use ELO system to select question pair
	eloSystem := elo.New(tournament)
	q1ID, q2ID, err := eloSystem.SelectPair()
	if err != nil {
		return nil, fmt.Errorf("failed to select question pair: %w", err)
	}
	
	// Get the actual questions from the database
	questions, err := s.questionRepo.FindByIDs([]int{q1ID, q2ID})
	if err != nil {
		return nil, fmt.Errorf("failed to find questions by IDs %v: %w", []int{q1ID, q2ID}, err)
	}
	
	if len(questions) != 2 {
		return nil, fmt.Errorf("expected 2 questions, got %d", len(questions))
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
		return 0, fmt.Errorf("failed to find active tournament: %w", err)
	}
	return tournament.QuestionsCount, nil
}