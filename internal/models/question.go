package models

import (
	"database/sql"
	"fmt"
	"questions-vote/internal/db"
)

// QuestionRepository handles question database operations
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new question repository
func NewQuestionRepository() *QuestionRepository {
	return &QuestionRepository{
		db: db.GetDB(),
	}
}

// FindByIDs retrieves questions by their IDs
func (r *QuestionRepository) FindByIDs(ids []int) ([]*Question, error) {
	if len(ids) == 0 {
		return []*Question{}, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT q.id, q.question, q.answer, q.comment, q.accepted_answer, 
		       q.handout_str, q.source, i.data as image_data 
		FROM questions q
		LEFT JOIN images i ON q.id = i.question_id
		WHERE q.id IN (%s)
	`, fmt.Sprintf("%s", placeholders[0]))
	
	for i := 1; i < len(placeholders); i++ {
		query = fmt.Sprintf("%s,%s", query[:len(query)-1], placeholders[i]) + ")"
	}

	// Rebuild query properly
	query = fmt.Sprintf(`
		SELECT q.id, q.question, q.answer, q.comment, q.accepted_answer, 
		       q.handout_str, q.source, i.data as image_data 
		FROM questions q
		LEFT JOIN images i ON q.id = i.question_id
		WHERE q.id IN (%s)
	`, joinStrings(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []*Question
	for rows.Next() {
		q := &Question{}
		var imageData sql.NullString
		err := rows.Scan(
			&q.ID,
			&q.Question,
			&q.Answer,
			&q.Comment,
			&q.AcceptedAnswer,
			&q.HandoutStr,
			&q.Source,
			&imageData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		if imageData.Valid {
			q.HandoutImg = imageData.String
		}

		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return questions, nil
}

// GetQuestionIDsForYear returns question IDs for a specific year
func (r *QuestionRepository) GetQuestionIDsForYear(year int) ([]int, error) {
	startTimestamp := int64(year * 365 * 24 * 3600) // Simplified timestamp calculation
	endTimestamp := int64((year + 1) * 365 * 24 * 3600)

	query := `
		SELECT id FROM questions 
		WHERE package_id IN (
			SELECT gotquestions_id FROM packages 
			WHERE end_date BETWEEN ? AND ?
		)
		AND is_incorrect = 0
	`

	rows, err := r.db.Query(query, startTimestamp, endTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to query question IDs: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan question ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}