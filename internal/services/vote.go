package services

import (
	"log"
	"questions-vote/internal/models"
)

// VoteService handles voting operations
type VoteService struct {
	voteRepo       *models.VoteRepository
	tournamentRepo *models.TournamentRepository
}

// NewVoteService creates a new vote service
func NewVoteService() *VoteService {
	return &VoteService{
		voteRepo:       models.NewVoteRepository(),
		tournamentRepo: models.NewTournamentRepository(),
	}
}

// SaveVote saves a user's vote
func (s *VoteService) SaveVote(userID int64, question1ID, question2ID int, selectedID *int) error {
	// Get active tournament
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		log.Printf("Failed to get active tournament: %v", err)
		return err
	}
	
	// Save the vote
	err = s.voteRepo.Create(userID, question1ID, question2ID, tournament.ID, selectedID)
	if err != nil {
		log.Printf("Failed to save vote: %v", err)
		return err
	}
	
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
	// TODO: Implement database lookup for actual stats
	// This should query tournament_questions table to get wins/matches
	return []models.QuestionStats{
		{Wins: 5, Matches: 10},
		{Wins: 3, Matches: 8},
	}, nil
}