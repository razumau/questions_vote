package elo

import (
	"fmt"
	"math"
	"questions-vote/internal/models"
	"sync"
)

var (
	retries int
	mu      sync.Mutex
)

// ELO implements the ELO rating system for tournament questions
type ELO struct {
	TournamentID           int
	InitialK               float64
	MinimumK               float64
	StdDevMultiplier       float64
	InitialPhaseMatches    int
	TransitionPhaseMatches int
	TopN                   int
	BandSize               int
	
	tournamentQuestionRepo *models.TournamentQuestionRepository
}

// New creates a new ELO instance from a tournament
func New(tournament *models.Tournament) *ELO {
	return &ELO{
		TournamentID:           tournament.ID,
		InitialK:               tournament.InitialK,
		MinimumK:               tournament.MinimumK,
		StdDevMultiplier:       tournament.StdDevMultiplier,
		InitialPhaseMatches:    tournament.InitialPhaseMatches,
		TransitionPhaseMatches: tournament.TransitionPhaseMatches,
		TopN:                   tournament.TopN,
		BandSize:               tournament.BandSize,
		tournamentQuestionRepo: models.NewTournamentQuestionRepository(),
	}
}

// SelectPair selects two questions for comparison based on ELO algorithm
func (e *ELO) SelectPair() (int, int, error) {
	mu.Lock()
	defer mu.Unlock()
	
	return e.selectPairInternal()
}

func (e *ELO) selectPairInternal() (int, int, error) {
	unqualifiedCount, err := e.tournamentQuestionRepo.CountUnqualifiedQuestions(e.TournamentID, e.InitialPhaseMatches)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count unqualified questions: %w", err)
	}
	
	var first, second int
	
	if unqualifiedCount > 1 {
		first, second, err = e.selectTwoUnqualified()
	} else if unqualifiedCount == 1 {
		first, err = e.selectUnqualified()
		if err != nil {
			return 0, 0, err
		}
		second, err = e.selectAny()
	} else {
		first, second, err = e.selectTwoQualified()
	}
	
	if err != nil {
		return 0, 0, err
	}
	
	if first == second {
		retries++
		return e.selectPairInternal()
	}
	
	return first, second, nil
}

// selectTwoUnqualified selects two unqualified questions
func (e *ELO) selectTwoUnqualified() (int, int, error) {
	maxMatches := e.InitialPhaseMatches - 1
	
	first, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, 0, math.MaxFloat64, maxMatches)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get first unqualified question: %w", err)
	}
	
	second, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, 0, math.MaxFloat64, maxMatches)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get second unqualified question: %w", err)
	}
	
	return first.QuestionID, second.QuestionID, nil
}

// selectUnqualified selects one unqualified question
func (e *ELO) selectUnqualified() (int, error) {
	maxMatches := e.InitialPhaseMatches - 1
	
	question, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, 0, math.MaxFloat64, maxMatches)
	if err != nil {
		return 0, fmt.Errorf("failed to get unqualified question: %w", err)
	}
	
	return question.QuestionID, nil
}

// selectAny selects any question
func (e *ELO) selectAny() (int, error) {
	question, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, 0, math.MaxFloat64, math.MaxInt32)
	if err != nil {
		return 0, fmt.Errorf("failed to get any question: %w", err)
	}
	
	return question.QuestionID, nil
}

// selectTwoQualified selects two qualified questions above threshold
func (e *ELO) selectTwoQualified() (int, int, error) {
	threshold, err := e.CalculateThreshold()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate threshold: %w", err)
	}
	
	first, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, threshold, math.MaxFloat64, math.MaxInt32)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get first qualified question: %w", err)
	}
	
	second, err := e.tournamentQuestionRepo.GetRandomQuestion(e.TournamentID, threshold, math.MaxFloat64, math.MaxInt32)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get second qualified question: %w", err)
	}
	
	return first.QuestionID, second.QuestionID, nil
}

// calculateKFactor calculates the K-factor for a tournament question
func (e *ELO) calculateKFactor(tq *models.TournamentQuestion) float64 {
	if tq.Matches < e.InitialPhaseMatches {
		return e.InitialK
	} else if tq.Matches < e.TransitionPhaseMatches {
		return e.InitialK / 2
	} else {
		return e.MinimumK
	}
}

// RecordWinner records the winner and updates ELO ratings
func (e *ELO) RecordWinner(winnerID, loserID int) error {
	winner, err := e.tournamentQuestionRepo.Find(e.TournamentID, winnerID)
	if err != nil {
		return fmt.Errorf("failed to find winner: %w", err)
	}
	
	loser, err := e.tournamentQuestionRepo.Find(e.TournamentID, loserID)
	if err != nil {
		return fmt.Errorf("failed to find loser: %w", err)
	}
	
	// Update match counts
	winner.Matches++
	loser.Matches++
	winner.Wins++
	
	// Calculate expected winner probability
	expectedWinner := 1.0 / (1.0 + math.Pow(10, (loser.Rating-winner.Rating)/400))
	
	// Calculate K-factors
	winnerK := e.calculateKFactor(winner)
	loserK := e.calculateKFactor(loser)
	kFactor := (winnerK + loserK) / 2
	
	// Calculate rating change
	ratingChange := kFactor * (1 - expectedWinner)
	
	// Update ratings
	winner.Rating += ratingChange
	loser.Rating -= ratingChange
	
	// Save changes
	err = e.tournamentQuestionRepo.Save(winner)
	if err != nil {
		return fmt.Errorf("failed to save winner: %w", err)
	}
	
	err = e.tournamentQuestionRepo.Save(loser)
	if err != nil {
		return fmt.Errorf("failed to save loser: %w", err)
	}
	
	return nil
}

// GetTopItems returns the top N questions by rating
func (e *ELO) GetTopItems(n int) ([]*models.TournamentQuestion, error) {
	return e.tournamentQuestionRepo.GetTopQuestions(e.TournamentID, n)
}

// CalculateThreshold calculates the rating threshold for question selection
func (e *ELO) CalculateThreshold() (float64, error) {
	ratingsCount, stdDev, err := e.tournamentQuestionRepo.GetStatsForQualified(e.TournamentID, e.InitialPhaseMatches)
	if err != nil {
		return 0, fmt.Errorf("failed to get stats for qualified: %w", err)
	}
	
	if ratingsCount < e.TopN {
		return math.Inf(-1), nil
	}
	
	topNThreshold, err := e.tournamentQuestionRepo.GetRatingAtPosition(e.TournamentID, e.TopN)
	if err != nil {
		return 0, fmt.Errorf("failed to get rating at position: %w", err)
	}
	
	return topNThreshold - (e.StdDevMultiplier * stdDev), nil
}

// GetQuestionsStats returns statistics for two questions
func (e *ELO) GetQuestionsStats(q1ID, q2ID int) ([]models.QuestionStats, error) {
	q1, err := e.tournamentQuestionRepo.Find(e.TournamentID, q1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find question 1: %w", err)
	}
	
	q2, err := e.tournamentQuestionRepo.Find(e.TournamentID, q2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find question 2: %w", err)
	}
	
	return []models.QuestionStats{
		{Wins: q1.Wins, Matches: q1.Matches},
		{Wins: q2.Wins, Matches: q2.Matches},
	}, nil
}

// GetStatistics returns comprehensive tournament statistics
func (e *ELO) GetStatistics() (map[string]interface{}, error) {
	mu.Lock()
	currentRetries := retries
	mu.Unlock()
	
	threshold, err := e.CalculateThreshold()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate threshold: %w", err)
	}
	
	aboveThresholdCount, err := e.tournamentQuestionRepo.CountQuestionsAboveThreshold(e.TournamentID, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to count questions above threshold: %w", err)
	}
	
	ratingDistribution, err := e.tournamentQuestionRepo.GetRatingDistribution(e.TournamentID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	
	unqualifiedCount, err := e.tournamentQuestionRepo.CountUnqualifiedQuestions(e.TournamentID, e.InitialPhaseMatches)
	if err != nil {
		return nil, fmt.Errorf("failed to count unqualified questions: %w", err)
	}
	
	totalMatches, totalWins, err := e.tournamentQuestionRepo.GetMatchCounts(e.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match counts: %w", err)
	}
	
	return map[string]interface{}{
		"current_threshold":  threshold,
		"above_threshold":    aboveThresholdCount,
		"unqualified":        unqualifiedCount,
		"distribution":       ratingDistribution,
		"retries":            currentRetries,
		"total_matches":      totalMatches,
		"total_wins":         totalWins,
	}, nil
}

// GetRetryCount returns the current retry count (for testing)
func GetRetryCount() int {
	mu.Lock()
	defer mu.Unlock()
	return retries
}

// ResetRetryCount resets the retry count (for testing)
func ResetRetryCount() {
	mu.Lock()
	defer mu.Unlock()
	retries = 0
}