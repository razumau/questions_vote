package models

import (
	"database/sql"
	"fmt"
	"questions-vote/internal/db"
)

// TournamentRepository handles tournament database operations
type TournamentRepository struct {
	db *sql.DB
}

// NewTournamentRepository creates a new tournament repository
func NewTournamentRepository() *TournamentRepository {
	return &TournamentRepository{
		db: db.GetDB(),
	}
}

// FindActiveTournament retrieves the currently active tournament
func (r *TournamentRepository) FindActiveTournament() (*Tournament, error) {
	query := `
		SELECT id, title, initial_k, minimum_k, std_dev_multiplier, 
		       initial_phase_matches, transition_phase_matches, top_n, 
		       questions_count, band_size
		FROM tournaments 
		WHERE state = 1
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active tournament: %w", err)
	}
	defer rows.Close()
	
	var tournaments []*Tournament
	for rows.Next() {
		t := &Tournament{}
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.InitialK,
			&t.MinimumK,
			&t.StdDevMultiplier,
			&t.InitialPhaseMatches,
			&t.TransitionPhaseMatches,
			&t.TopN,
			&t.QuestionsCount,
			&t.BandSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tournament: %w", err)
		}
		tournaments = append(tournaments, t)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tournaments: %w", err)
	}
	
	if len(tournaments) == 0 {
		return nil, fmt.Errorf("no active tournament found")
	}
	
	if len(tournaments) > 1 {
		return nil, fmt.Errorf("found %d active tournaments, expected exactly 1", len(tournaments))
	}
	
	return tournaments[0], nil
}

// ListActiveTournaments returns all active tournaments
func (r *TournamentRepository) ListActiveTournaments() ([]*Tournament, error) {
	query := `
		SELECT id, title, initial_k, minimum_k, std_dev_multiplier, 
		       initial_phase_matches, transition_phase_matches, top_n, 
		       questions_count, band_size
		FROM tournaments 
		WHERE state = 1
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active tournaments: %w", err)
	}
	defer rows.Close()
	
	var tournaments []*Tournament
	for rows.Next() {
		t := &Tournament{}
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.InitialK,
			&t.MinimumK,
			&t.StdDevMultiplier,
			&t.InitialPhaseMatches,
			&t.TransitionPhaseMatches,
			&t.TopN,
			&t.QuestionsCount,
			&t.BandSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tournament: %w", err)
		}
		tournaments = append(tournaments, t)
	}
	
	return tournaments, rows.Err()
}

// StartTournament activates a tournament by title
func (r *TournamentRepository) StartTournament(title string) error {
	query := `UPDATE tournaments SET state = 1 WHERE title = ?`
	_, err := r.db.Exec(query, title)
	if err != nil {
		return fmt.Errorf("failed to start tournament: %w", err)
	}
	return nil
}