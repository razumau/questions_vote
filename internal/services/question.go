package services

import "questions-vote/internal/models"

// QuestionService handles question-related operations
type QuestionService struct {
	// TODO: Add database connection
}

// NewQuestionService creates a new question service
func NewQuestionService() *QuestionService {
	return &QuestionService{}
}

// GetQuestions returns a pair of questions for voting
func (s *QuestionService) GetQuestions() ([]*models.Question, error) {
	// TODO: Implement ELO-based question selection
	// For now, return mock data
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

// FindQuestions finds questions by IDs
func (s *QuestionService) FindQuestions(ids []int) ([]*models.Question, error) {
	// TODO: Implement database lookup
	return s.GetQuestions()
}