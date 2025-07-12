package models

import (
	"database/sql"
	"fmt"
	"questions-vote/internal/db"
	"strings"
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
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
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
	`, strings.Join(placeholders, ","))

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
	startTimestamp := int64(year * 365 * 24 * 3600)
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

// HasQuestionsFromPackage checks if a package has any questions
func (r *QuestionRepository) HasQuestionsFromPackage(packageID int) (bool, error) {
	query := `SELECT 1 FROM questions WHERE package_id = ? LIMIT 1`

	var exists int
	err := r.db.QueryRow(query, packageID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if package has questions: %w", err)
	}

	return true, nil
}

// BuildQuestionFromDict creates a question from API response data
func BuildQuestionFromDict(questionDict map[string]interface{}, packageID int) (*Question, error) {
	id, ok := questionDict["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid id in question data")
	}

	text, ok := questionDict["text"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid text in question data")
	}

	answer, ok := questionDict["answer"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid answer in question data")
	}

	var zachet, comment, razdatkaText, razdatkaPic, source string
	var complexity float64
	var takenDown bool
	var authorID *int

	if val, exists := questionDict["zachet"]; exists && val != nil {
		zachet, _ = val.(string)
	}

	if val, exists := questionDict["comment"]; exists && val != nil {
		comment, _ = val.(string)
	}

	if val, exists := questionDict["razdatkaText"]; exists && val != nil {
		razdatkaText, _ = val.(string)
	}

	if val, exists := questionDict["razdatkaPic"]; exists && val != nil {
		razdatkaPic, _ = val.(string)
	}

	if val, exists := questionDict["source"]; exists && val != nil {
		source, _ = val.(string)
	}

	if val, exists := questionDict["complexity"]; exists && val != nil {
		complexity, _ = val.(float64)
	}

	if val, exists := questionDict["takenDown"]; exists && val != nil {
		takenDown, _ = val.(bool)
	}

	if authorsInterface, exists := questionDict["authors"]; exists && authorsInterface != nil {
		if authors, ok := authorsInterface.([]any); ok && len(authors) > 0 {
			if author, ok := authors[0].(map[string]any); ok {
				if idVal, exists := author["id"]; exists && idVal != nil {
					if idFloat, ok := idVal.(float64); ok {
						id := int(idFloat)
						authorID = &id
					}
				}
			}
		}
	}

	return &Question{
		GotQuestionsID: int(id),
		Question:       text,
		Answer:         answer,
		AcceptedAnswer: zachet,
		Comment:        comment,
		HandoutStr:     razdatkaText,
		HandoutImg:     razdatkaPic,
		Source:         source,
		AuthorID:       authorID,
		PackageID:      &packageID,
		Difficulty:     &complexity,
		IsIncorrect:    &takenDown,
	}, nil
}

// GetQuestionIDsFromPackage returns question IDs for a specific package
func (r *QuestionRepository) GetQuestionIDsFromPackage(packageID int) ([]int, error) {
	query := `SELECT id FROM questions WHERE package_id = ?`

	rows, err := r.db.Query(query, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questionIDs []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question ID: %w", err)
		}
		questionIDs = append(questionIDs, id)
	}

	return questionIDs, rows.Err()
}
