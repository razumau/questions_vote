package models

import (
	"database/sql"
	"fmt"
	"questions-vote/internal/db"
	"time"
)

// Package represents a question package from gotquestions.online
type Package struct {
	ID             int       `json:"id"`
	GotQuestionsID int       `json:"gotquestions_id"`
	Title          string    `json:"title"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	QuestionsCount int       `json:"questions_count"`
}

// PackageRepository handles package database operations
type PackageRepository struct {
	db *sql.DB
}

// NewPackageRepository creates a new package repository
func NewPackageRepository() *PackageRepository {
	return &PackageRepository{
		db: db.GetDB(),
	}
}

// BuildPackageFromDict creates a package from API response data
func BuildPackageFromDict(packageDict map[string]any) (*Package, error) {
	id, ok := packageDict["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid id in package data")
	}

	title, ok := packageDict["title"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid title in package data")
	}

	startDateStr, ok := packageDict["startDate"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid startDate in package data")
	}

	endDateStr, ok := packageDict["endDate"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid endDate in package data")
	}

	questionsCount, ok := packageDict["questions"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid questions count in package data")
	}

	startDate, err := time.ParseInLocation("2006-01-02T15:04:05", startDateStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate, err := time.ParseInLocation("2006-01-02T15:04:05", endDateStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end date: %w", err)
	}

	return &Package{
		GotQuestionsID: int(id),
		Title:          title,
		StartDate:      startDate,
		EndDate:        endDate,
		QuestionsCount: int(questionsCount),
	}, nil
}

// Insert inserts a package into the database
func (r *PackageRepository) Insert(pkg *Package) error {
	if r.Exists(pkg.GotQuestionsID) {
		return nil // Already exists, skip
	}

	query := `
		INSERT INTO packages (gotquestions_id, title, start_date, end_date, questions_count)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, pkg.GotQuestionsID, pkg.Title, pkg.StartDate, pkg.EndDate, pkg.QuestionsCount)
	if err != nil {
		return fmt.Errorf("failed to insert package: %w", err)
	}

	return nil
}

// Exists checks if a package exists by gotquestions_id
func (r *PackageRepository) Exists(gotQuestionsID int) bool {
	query := `SELECT 1 FROM packages WHERE gotquestions_id = ? LIMIT 1`

	var exists int
	err := r.db.QueryRow(query, gotQuestionsID).Scan(&exists)
	return err == nil
}

// FindByGotQuestionsID finds a package by its gotquestions.online ID
func (r *PackageRepository) FindByGotQuestionsID(gotQuestionsID int) (*Package, error) {
	query := `
		SELECT id, gotquestions_id, title, start_date, end_date, questions_count
		FROM packages 
		WHERE gotquestions_id = ?
	`

	pkg := &Package{}
	err := r.db.QueryRow(query, gotQuestionsID).Scan(
		&pkg.ID, &pkg.GotQuestionsID, &pkg.Title,
		&pkg.StartDate, &pkg.EndDate, &pkg.QuestionsCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package not found")
		}
		return nil, fmt.Errorf("failed to find package: %w", err)
	}

	return pkg, nil
}

// GetPackagesByYear returns all packages for a specific year
func (r *PackageRepository) GetPackagesByYear(year int) ([]*Package, error) {
	query := `
		SELECT id, gotquestions_id, title, start_date, end_date, questions_count
		FROM packages 
		WHERE strftime('%Y', end_date) = ?
		ORDER BY end_date
	`

	rows, err := r.db.Query(query, fmt.Sprintf("%04d", year))
	if err != nil {
		return nil, fmt.Errorf("failed to query packages by year: %w", err)
	}
	defer rows.Close()

	var packages []*Package
	for rows.Next() {
		pkg := &Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.GotQuestionsID, &pkg.Title,
			&pkg.StartDate, &pkg.EndDate, &pkg.QuestionsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	return packages, rows.Err()
}

// GetAllPackages returns all packages
func (r *PackageRepository) GetAllPackages() ([]*Package, error) {
	query := `
		SELECT id, gotquestions_id, title, start_date, end_date, questions_count
		FROM packages 
		ORDER BY end_date DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all packages: %w", err)
	}
	defer rows.Close()

	var packages []*Package
	for rows.Next() {
		pkg := &Package{}
		err := rows.Scan(
			&pkg.ID, &pkg.GotQuestionsID, &pkg.Title,
			&pkg.StartDate, &pkg.EndDate, &pkg.QuestionsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	return packages, rows.Err()
}
