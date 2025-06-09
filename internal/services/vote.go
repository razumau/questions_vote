package services

import (
	"fmt"
	"log"
	"questions-vote/internal/elo"
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
		// Update ELO ratings
		var loserID int
		if *selectedID == question2ID {
			loserID = question1ID
		} else {
			loserID = question2ID
		}
		
		// Use ELO system to record the winner
		eloSystem := elo.New(tournament)
		err = eloSystem.RecordWinner(*selectedID, loserID)
		if err != nil {
			log.Printf("Failed to record ELO winner: %v", err)
			// Don't return error - vote was still saved
		} else {
			log.Printf("Recorded ELO winner: %d, loser: %d", *selectedID, loserID)
		}
	}
	
	return nil
}

// GetQuestionStats returns statistics for questions using ELO system
func (s *VoteService) GetQuestionStats(question1ID, question2ID int) ([]models.QuestionStats, error) {
	// Get active tournament
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		return nil, fmt.Errorf("failed to find active tournament: %w", err)
	}
	
	// Use ELO system to get question stats
	eloSystem := elo.New(tournament)
	stats, err := eloSystem.GetQuestionsStats(question1ID, question2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get question stats from ELO: %w", err)
	}
	
	return stats, nil
}