package models

import (
	"database/sql"
	"fmt"
	"questions-vote/internal/db"
	"time"
)

// VoteRepository handles vote database operations
type VoteRepository struct {
	db *sql.DB
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository() *VoteRepository {
	return &VoteRepository{
		db: db.GetDB(),
	}
}

// Create inserts a new vote into the database
func (r *VoteRepository) Create(userID int64, question1ID, question2ID, tournamentID int, selectedID *int) error {
	query := `
		INSERT INTO votes (user_id, question1_id, question2_id, tournament_id, selected_id, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		userID,
		question1ID,
		question2ID,
		tournamentID,
		selectedID,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}

	return nil
}

// GetVoteCount returns the total number of votes for a user
func (r *VoteRepository) GetVoteCount(userID int64, tournamentID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM votes 
		WHERE user_id = ? AND tournament_id = ?
	`

	var count int
	err := r.db.QueryRow(query, userID, tournamentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get vote count: %w", err)
	}

	return count, nil
}

// GetRecentVotes returns recent votes for a user
func (r *VoteRepository) GetRecentVotes(userID int64, tournamentID int, limit int) ([]*Vote, error) {
	query := `
		SELECT id, user_id, question1_id, question2_id, tournament_id, selected_id, timestamp
		FROM votes 
		WHERE user_id = ? AND tournament_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, tournamentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent votes: %w", err)
	}
	defer rows.Close()

	var votes []*Vote
	for rows.Next() {
		v := &Vote{}
		err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.Question1ID,
			&v.Question2ID,
			&v.TournamentID,
			&v.SelectedID,
			&v.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vote: %w", err)
		}
		votes = append(votes, v)
	}

	return votes, rows.Err()
}
