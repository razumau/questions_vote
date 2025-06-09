package models

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"questions-vote/internal/db"
)

// TournamentQuestion represents a question in a tournament with ELO rating
type TournamentQuestion struct {
	TournamentID int     `json:"tournament_id"`
	QuestionID   int     `json:"question_id"`
	Rating       float64 `json:"rating"`
	Matches      int     `json:"matches"`
	Wins         int     `json:"wins"`
}

// TournamentQuestionRepository handles tournament question database operations
type TournamentQuestionRepository struct {
	db *sql.DB
}

// NewTournamentQuestionRepository creates a new tournament question repository
func NewTournamentQuestionRepository() *TournamentQuestionRepository {
	return &TournamentQuestionRepository{
		db: db.GetDB(),
	}
}

// CreateTournamentQuestions creates tournament questions for a tournament
func (r *TournamentQuestionRepository) CreateTournamentQuestions(tournamentID int, questionIDs []int, initialRating float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO tournament_questions (tournament_id, question_id, rating, matches, wins) 
		VALUES (?, ?, ?, 0, 0)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, questionID := range questionIDs {
		_, err = stmt.Exec(tournamentID, questionID, initialRating)
		if err != nil {
			return fmt.Errorf("failed to insert tournament question: %w", err)
		}
	}

	return tx.Commit()
}

// GetRandomQuestion returns a random question matching the criteria
func (r *TournamentQuestionRepository) GetRandomQuestion(tournamentID int, minRating, maxRating float64, maxMatches int) (*TournamentQuestion, error) {
	// First get the count of matching questions
	countQuery := `
		SELECT COUNT(*) 
		FROM tournament_questions 
		WHERE tournament_id = ? 
		AND rating BETWEEN ? AND ? 
		AND matches <= ?
	`
	
	var count int
	err := r.db.QueryRow(countQuery, tournamentID, minRating, maxRating, maxMatches).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count questions: %w", err)
	}
	
	if count == 0 {
		return nil, fmt.Errorf("no questions found matching criteria")
	}
	
	// Get a random offset
	offset := rand.Intn(count)
	
	// Get the question at that offset
	query := `
		SELECT question_id, rating, matches, wins
		FROM tournament_questions
		WHERE tournament_id = ? 
		AND rating BETWEEN ? AND ?
		AND matches <= ?
		LIMIT 1 OFFSET ?
	`
	
	tq := &TournamentQuestion{TournamentID: tournamentID}
	err = r.db.QueryRow(query, tournamentID, minRating, maxRating, maxMatches, offset).Scan(
		&tq.QuestionID, &tq.Rating, &tq.Matches, &tq.Wins,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get random question: %w", err)
	}
	
	return tq, nil
}

// Find returns a tournament question by tournament and question ID
func (r *TournamentQuestionRepository) Find(tournamentID, questionID int) (*TournamentQuestion, error) {
	query := `
		SELECT rating, matches, wins
		FROM tournament_questions
		WHERE tournament_id = ? AND question_id = ?
	`
	
	tq := &TournamentQuestion{
		TournamentID: tournamentID,
		QuestionID:   questionID,
	}
	
	err := r.db.QueryRow(query, tournamentID, questionID).Scan(
		&tq.Rating, &tq.Matches, &tq.Wins,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tournament question not found")
		}
		return nil, fmt.Errorf("failed to find tournament question: %w", err)
	}
	
	return tq, nil
}

// Save updates a tournament question's rating, matches, and wins
func (r *TournamentQuestionRepository) Save(tq *TournamentQuestion) error {
	query := `
		UPDATE tournament_questions
		SET rating = ?, matches = ?, wins = ?
		WHERE tournament_id = ? AND question_id = ?
	`
	
	_, err := r.db.Exec(query, tq.Rating, tq.Matches, tq.Wins, tq.TournamentID, tq.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to save tournament question: %w", err)
	}
	
	return nil
}

// CountUnqualifiedQuestions returns count of questions with matches < qualification cutoff
func (r *TournamentQuestionRepository) CountUnqualifiedQuestions(tournamentID, qualificationCutoff int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tournament_questions
		WHERE tournament_id = ? AND matches < ?
	`
	
	var count int
	err := r.db.QueryRow(query, tournamentID, qualificationCutoff).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unqualified questions: %w", err)
	}
	
	return count, nil
}

// GetStatsForQualified returns count and standard deviation of qualified questions
func (r *TournamentQuestionRepository) GetStatsForQualified(tournamentID, initialPhaseMatches int) (int, float64, error) {
	// First get count and sum of ratings and sum of squares
	query := `
		SELECT COUNT(rating), 
		       COALESCE(SUM(rating), 0) as sum_rating,
		       COALESCE(SUM(rating * rating), 0) as sum_squares
		FROM tournament_questions
		WHERE tournament_id = ? AND matches >= ?
	`
	
	var count int
	var sumRating, sumSquares float64
	err := r.db.QueryRow(query, tournamentID, initialPhaseMatches).Scan(&count, &sumRating, &sumSquares)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get stats for qualified: %w", err)
	}
	
	var stdDev float64
	if count > 1 {
		// Calculate standard deviation using the formula: sqrt((sum_squares - sum^2/n) / (n-1))
		mean := sumRating / float64(count)
		variance := (sumSquares - sumRating*mean) / float64(count-1)
		if variance >= 0 {
			stdDev = math.Sqrt(variance)
		}
	}
	
	return count, stdDev, nil
}

// GetRatingAtPosition returns the rating at a specific position (0-indexed)
func (r *TournamentQuestionRepository) GetRatingAtPosition(tournamentID, position int) (float64, error) {
	query := `
		SELECT rating
		FROM tournament_questions
		WHERE tournament_id = ?
		ORDER BY rating DESC
		LIMIT 1 OFFSET ?
	`
	
	var rating float64
	err := r.db.QueryRow(query, tournamentID, position).Scan(&rating)
	if err != nil {
		return 0, fmt.Errorf("failed to get rating at position: %w", err)
	}
	
	return rating, nil
}

// CountQuestionsAboveThreshold returns count of questions above a rating threshold
func (r *TournamentQuestionRepository) CountQuestionsAboveThreshold(tournamentID int, threshold float64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tournament_questions
		WHERE tournament_id = ? AND rating >= ?
	`
	
	var count int
	err := r.db.QueryRow(query, tournamentID, threshold).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count questions above threshold: %w", err)
	}
	
	return count, nil
}

// GetTopQuestions returns the top N questions by rating
func (r *TournamentQuestionRepository) GetTopQuestions(tournamentID, n int) ([]*TournamentQuestion, error) {
	query := `
		SELECT question_id, rating, matches, wins
		FROM tournament_questions
		WHERE tournament_id = ?
		ORDER BY rating DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, tournamentID, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get top questions: %w", err)
	}
	defer rows.Close()
	
	var questions []*TournamentQuestion
	for rows.Next() {
		tq := &TournamentQuestion{TournamentID: tournamentID}
		err := rows.Scan(&tq.QuestionID, &tq.Rating, &tq.Matches, &tq.Wins)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tournament question: %w", err)
		}
		questions = append(questions, tq)
	}
	
	return questions, rows.Err()
}

// GetMatchCounts returns total matches and wins for a tournament
func (r *TournamentQuestionRepository) GetMatchCounts(tournamentID int) (int, int, error) {
	query := `
		SELECT COALESCE(SUM(matches), 0), COALESCE(SUM(wins), 0)
		FROM tournament_questions
		WHERE tournament_id = ?
	`
	
	var totalMatches, totalWins int
	err := r.db.QueryRow(query, tournamentID).Scan(&totalMatches, &totalWins)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get match counts: %w", err)
	}
	
	return totalMatches, totalWins, nil
}

// GetRatingDistribution returns rating distribution in bins
func (r *TournamentQuestionRepository) GetRatingDistribution(tournamentID, binSize int) (map[int]int, error) {
	query := `
		SELECT ROUND(rating / ?) * ? as bin, COUNT(*) as count
		FROM tournament_questions
		WHERE tournament_id = ?
		GROUP BY bin
		ORDER BY bin
	`
	
	rows, err := r.db.Query(query, binSize, binSize, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	defer rows.Close()
	
	distribution := make(map[int]int)
	for rows.Next() {
		var bin, count int
		err := rows.Scan(&bin, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating distribution: %w", err)
		}
		distribution[bin] = count
	}
	
	return distribution, rows.Err()
}