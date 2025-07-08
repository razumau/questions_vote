package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"questions-vote/internal/db"
	"questions-vote/internal/models"
	"time"
)

func main() {
	var (
		command         = flag.String("command", "", "Command to run: create-tournament")
		earliestDate    = flag.String("earliest-date", "", "Earliest package date (YYYY-MM-DD)")
		lastDate        = flag.String("last-date", "", "Last package date (YYYY-MM-DD)")
		tournamentTitle = flag.String("title", "", "Tournament title")
	)
	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	err := db.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	switch *command {
	case "create-tournament":
		if *earliestDate == "" || *lastDate == "" || *tournamentTitle == "" {
			log.Fatal("earliest-date, last-date, and title are required for create-tournament command")
		}
		err = runCreateTournament(*earliestDate, *lastDate, *tournamentTitle)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

func printUsage() {
	fmt.Println("Questions Vote Tournament Manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tournament_manager -command=create-tournament -earliest-date=YYYY-MM-DD -last-date=YYYY-MM-DD -title=\"Tournament Name\"")
	fmt.Println("    Creates a new tournament with questions from packages between the specified dates")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tournament_manager -command=create-tournament -earliest-date=2023-01-01 -last-date=2023-12-31 -title=\"2023 Tournament\"")
}

func runCreateTournament(earliestDateStr, lastDateStr, title string) error {
	earliestDate, err := time.Parse("2006-01-02", earliestDateStr)
	if err != nil {
		return fmt.Errorf("invalid earliest date format: %w", err)
	}

	lastDate, err := time.Parse("2006-01-02", lastDateStr)
	if err != nil {
		return fmt.Errorf("invalid last date format: %w", err)
	}

	if !lastDate.After(earliestDate) {
		return fmt.Errorf("last date must be after earliest date")
	}

	log.Printf("Creating tournament '%s' with packages from %s to %s", title, earliestDateStr, lastDateStr)

	tournament := &models.Tournament{
		Name:                   title,
		InitialK:               64.0,
		MinimumK:               16.0,
		StdDevMultiplier:       1.5,
		InitialPhaseMatches:    5,
		TransitionPhaseMatches: 10,
		TopN:                   100,
		BandSize:               200,
	}

	tournamentRepo := models.NewTournamentRepository()
	tournamentID, err := createTournament(tournamentRepo, tournament)
	if err != nil {
		return fmt.Errorf("failed to create tournament: %w", err)
	}

	log.Printf("Tournament created with ID: %d", tournamentID)

	packageRepo := models.NewPackageRepository()
	packages, err := getPackagesByDateRange(packageRepo, earliestDate, lastDate)
	if err != nil {
		return fmt.Errorf("failed to get packages: %w", err)
	}

	if len(packages) == 0 {
		log.Println("No packages found in the specified date range")
		return nil
	}

	log.Printf("Found %d packages in date range", len(packages))

	questionRepo := models.NewQuestionRepository()
	var allQuestionIDs []int

	for _, pkg := range packages {
		questionIDs, err := getQuestionIDsFromPackage(questionRepo, pkg.GotQuestionsID)
		if err != nil {
			log.Printf("Warning: failed to get questions from package %d (%s): %v", pkg.GotQuestionsID, pkg.Title, err)
			continue
		}
		allQuestionIDs = append(allQuestionIDs, questionIDs...)
		log.Printf("Added %d questions from package %d (%s)", len(questionIDs), pkg.GotQuestionsID, pkg.Title)
	}

	if len(allQuestionIDs) == 0 {
		return fmt.Errorf("no questions found in packages within date range")
	}

	log.Printf("Total questions to add to tournament: %d", len(allQuestionIDs))

	tournamentQuestionRepo := models.NewTournamentQuestionRepository()
	err = tournamentQuestionRepo.CreateTournamentQuestions(tournamentID, allQuestionIDs, 1500.0)
	if err != nil {
		return fmt.Errorf("failed to create tournament questions: %w", err)
	}

	err = updateTournamentQuestionsCount(tournamentRepo, tournamentID, len(allQuestionIDs))
	if err != nil {
		return fmt.Errorf("failed to update tournament questions count: %w", err)
	}

	log.Printf("Tournament '%s' created successfully with %d questions", title, len(allQuestionIDs))
	return nil
}

func createTournament(repo *models.TournamentRepository, tournament *models.Tournament) (int, error) {
	// Note: This function needs to be implemented in tournament.go
	// For now, we'll do a direct database insert
	query := `
		INSERT INTO tournaments (title, initial_k, minimum_k, std_dev_multiplier, 
		                        initial_phase_matches, transition_phase_matches, top_n, 
		                        band_size, questions_count, state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0)
	`

	result, err := db.GetDB().Exec(query, tournament.Name, tournament.InitialK, tournament.MinimumK,
		tournament.StdDevMultiplier, tournament.InitialPhaseMatches, tournament.TransitionPhaseMatches,
		tournament.TopN, tournament.BandSize)
	if err != nil {
		return 0, fmt.Errorf("failed to insert tournament: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get tournament ID: %w", err)
	}

	return int(id), nil
}

func getPackagesByDateRange(repo *models.PackageRepository, startDate, endDate time.Time) ([]*models.Package, error) {
	// This function needs to be implemented in package.go
	// For now, we'll do a direct database query
	query := `
		SELECT id, gotquestions_id, title, start_date, end_date, questions_count
		FROM packages 
		WHERE start_date >= ? AND start_date <= ?
		ORDER BY start_date
	`

	rows, err := db.GetDB().Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query packages by date range: %w", err)
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
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

func getQuestionIDsFromPackage(repo *models.QuestionRepository, packageID int) ([]int, error) {
	// This function needs to be implemented in question.go
	// For now, we'll do a direct database query
	query := `SELECT id FROM questions WHERE package_id = ?`

	rows, err := db.GetDB().Query(query, packageID)
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

func updateTournamentQuestionsCount(repo *models.TournamentRepository, tournamentID, count int) error {
	query := `UPDATE tournaments SET questions_count = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, count, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to update tournament questions count: %w", err)
	}
	return nil
}
