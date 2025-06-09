package services

import (
	"log"
	"questions-vote/internal/models"
)

// VoteService handles voting operations
type VoteService struct {
	// TODO: Add database connection
}

// NewVoteService creates a new vote service
func NewVoteService() *VoteService {
	return &VoteService{}
}

// SaveVote saves a user's vote
func (s *VoteService) SaveVote(userID int64, question1ID, question2ID int, selectedID *int) error {
	// TODO: Implement database save
	log.Printf("Saving vote: user=%d, q1=%d, q2=%d, selected=%v", userID, question1ID, question2ID, selectedID)
	
	if selectedID != nil {
		// TODO: Update ELO ratings
		var loserID int
		if *selectedID == question2ID {
			loserID = question1ID
		} else {
			loserID = question2ID
		}
		log.Printf("Recording winner: %d, loser: %d", *selectedID, loserID)
	}
	
	return nil
}

// GetQuestionStats returns statistics for questions
func (s *VoteService) GetQuestionStats(question1ID, question2ID int) ([]models.QuestionStats, error) {
	// TODO: Implement database lookup
	return []models.QuestionStats{
		{Wins: 5, Matches: 10},
		{Wins: 3, Matches: 8},
	}, nil
}