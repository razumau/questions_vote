package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"questions-vote/internal/db"
	"questions-vote/internal/models"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		command         = flag.String("command", "", "Command to run: create-tournament, list-tournaments, activate-tournament, deactivate-tournament")
		earliestDate    = flag.String("earliest-date", "", "Earliest package date (YYYY-MM-DD)")
		lastDate        = flag.String("last-date", "", "Last package date (YYYY-MM-DD)")
		tournamentTitle = flag.String("title", "", "Tournament title")
		tournamentID    = flag.String("id", "", "Tournament ID")
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
	case "list-tournaments":
		err = runListTournaments()
	case "activate-tournament":
		if *tournamentID == "" {
			log.Fatal("id is required for activate-tournament command")
		}
		err = runActivateTournament(*tournamentID)
	case "deactivate-tournament":
		if *tournamentID == "" {
			log.Fatal("id is required for deactivate-tournament command")
		}
		err = runDeactivateTournament(*tournamentID)
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
	fmt.Println("  tournament_manager -command=list-tournaments")
	fmt.Println("    Lists all tournaments with their status")
	fmt.Println()
	fmt.Println("  tournament_manager -command=activate-tournament -id=TOURNAMENT_ID")
	fmt.Println("    Activates a tournament by ID")
	fmt.Println()
	fmt.Println("  tournament_manager -command=deactivate-tournament -id=TOURNAMENT_ID")
	fmt.Println("    Deactivates a tournament by ID")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tournament_manager -command=create-tournament -earliest-date=2023-01-01 -last-date=2023-12-31 -title=\"2023 Tournament\"")
	fmt.Println("  tournament_manager -command=list-tournaments")
	fmt.Println("  tournament_manager -command=activate-tournament -id=1")
	fmt.Println("  tournament_manager -command=deactivate-tournament -id=1")
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
	tournamentID, err := tournamentRepo.Create(tournament)
	if err != nil {
		return fmt.Errorf("failed to create tournament: %w", err)
	}

	log.Printf("Tournament created with ID: %d", tournamentID)

	packageRepo := models.NewPackageRepository()
	packages, err := packageRepo.GetPackagesByDateRange(earliestDate, lastDate)
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
		questionIDs, err := questionRepo.GetQuestionIDsFromPackage(pkg.GotQuestionsID)
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

	err = tournamentRepo.UpdateQuestionsCount(tournamentID, len(allQuestionIDs))
	if err != nil {
		return fmt.Errorf("failed to update tournament questions count: %w", err)
	}

	log.Printf("Tournament '%s' created successfully with %d questions", title, len(allQuestionIDs))
	return nil
}

func runListTournaments() error {
	tournamentRepo := models.NewTournamentRepository()
	tournaments, err := tournamentRepo.ListAllTournaments()
	if err != nil {
		return fmt.Errorf("failed to list tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		fmt.Println("No tournaments found")
		return nil
	}

	fmt.Printf("%-5s %-30s %-10s %-15s\n", "ID", "Name", "Active", "Questions")
	fmt.Println(strings.Repeat("-", 65))
	for _, tournament := range tournaments {
		status := "No"
		if tournament.Active {
			status = "Yes"
		}
		fmt.Printf("%-5d %-30s %-10s %-15d\n", tournament.ID, tournament.Name, status, tournament.QuestionsCount)
	}

	return nil
}

func runActivateTournament(tournamentIDStr string) error {
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("invalid tournament ID: %w", err)
	}

	tournamentRepo := models.NewTournamentRepository()
	err = tournamentRepo.ActivateTournament(tournamentID)
	if err != nil {
		return err
	}

	log.Printf("Tournament with ID %d activated successfully", tournamentID)
	return nil
}

func runDeactivateTournament(tournamentIDStr string) error {
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		return fmt.Errorf("invalid tournament ID: %w", err)
	}

	tournamentRepo := models.NewTournamentRepository()
	err = tournamentRepo.DeactivateTournament(tournamentID)
	if err != nil {
		return err
	}

	log.Printf("Tournament with ID %d deactivated successfully", tournamentID)
	return nil
}
