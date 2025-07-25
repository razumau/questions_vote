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
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		log.Printf("Failed to get active tournament: %v", err)
		return err
	}

	err = s.voteRepo.Create(userID, question1ID, question2ID, tournament.ID, selectedID)
	if err != nil {
		log.Printf("Failed to save vote: %v", err)
		return err
	}

	if selectedID != nil {
		log.Printf("Saving vote: user=%d, q1=%d, q2=%d, selected=%d", userID, question1ID, question2ID, *selectedID)
	} else {
		log.Printf("Saving vote: user=%d, q1=%d, q2=%d, selected=nil", userID, question1ID, question2ID)
	}

	if selectedID != nil {
		var loserID int
		if *selectedID == question2ID {
			loserID = question1ID
		} else {
			loserID = question2ID
		}

		eloSystem := elo.New(tournament)
		err = eloSystem.RecordWinner(*selectedID, loserID)
		if err != nil {
			log.Printf("Failed to record ELO winner: %v", err)
		} else {
			log.Printf("Recorded ELO winner: %d, loser: %d", *selectedID, loserID)
		}
	}

	return nil
}

// GetQuestionStats returns statistics for questions using ELO system
func (s *VoteService) GetQuestionStats(question1ID, question2ID int) ([]models.QuestionStats, error) {
	tournament, err := s.tournamentRepo.FindActiveTournament()
	if err != nil {
		return nil, fmt.Errorf("failed to find active tournament: %w", err)
	}

	eloSystem := elo.New(tournament)
	stats, err := eloSystem.GetQuestionsStats(question1ID, question2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get question stats from ELO: %w", err)
	}

	return stats, nil
}
