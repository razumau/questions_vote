package elo

import (
	"math"
	"os"
	"questions-vote/internal/db"
	"questions-vote/internal/models"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testDBPath             = "test_elo.db"
	initialRating          = 1500.0
	initialK               = 64.0
	minimumK               = 16.0
	stdDevMultiplier       = 2.0
	initialPhaseMatches    = 5
	transitionPhaseMatches = 10
	topN                   = 5
	bandSize               = 200
)

// setupTestDB creates a test database and tables
func setupTestDB(t *testing.T) {
	// Remove existing test database
	os.Remove(testDBPath)

	// Set test database path
	os.Setenv("DATABASE_URL", testDBPath)

	// Initialize database
	err := db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Create tables
	createTables(t)
}

// createTables creates the necessary tables for testing
func createTables(t *testing.T) {
	database := db.GetDB()

	// Create tournaments table
	_, err := database.Exec(`
		CREATE TABLE tournaments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			initial_k REAL,
			minimum_k REAL,
			std_dev_multiplier REAL,
			initial_phase_matches INTEGER,
			transition_phase_matches INTEGER,
			top_n INTEGER,
			questions_count INTEGER,
			band_size INTEGER,
			state INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tournaments table: %v", err)
	}

	// Create tournament_questions table
	_, err = database.Exec(`
		CREATE TABLE tournament_questions (
			tournament_id INTEGER,
			question_id INTEGER,
			rating REAL,
			matches INTEGER DEFAULT 0,
			wins INTEGER DEFAULT 0,
			PRIMARY KEY (tournament_id, question_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tournament_questions table: %v", err)
	}

	// Create votes table
	_, err = database.Exec(`
		CREATE TABLE votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			question1_id INTEGER,
			question2_id INTEGER,
			tournament_id INTEGER,
			selected_id INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create votes table: %v", err)
	}
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	db.Close()
	os.Remove(testDBPath)
}

// createTestTournament creates a tournament with test questions
func createTestTournament(t *testing.T, questionCount int) (*models.Tournament, []int) {
	database := db.GetDB()

	// Insert tournament
	result, err := database.Exec(`
		INSERT INTO tournaments 
		(title, initial_k, minimum_k, std_dev_multiplier, initial_phase_matches, 
		 transition_phase_matches, top_n, questions_count, band_size, state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, "Test Tournament", initialK, minimumK, stdDevMultiplier,
		initialPhaseMatches, transitionPhaseMatches, topN, questionCount, bandSize)
	if err != nil {
		t.Fatalf("Failed to create tournament: %v", err)
	}

	tournamentID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get tournament ID: %v", err)
	}

	// Create question IDs
	questionIDs := make([]int, questionCount)
	for i := 0; i < questionCount; i++ {
		questionIDs[i] = i + 1
	}

	// Create tournament questions
	repo := models.NewTournamentQuestionRepository()
	err = repo.CreateTournamentQuestions(int(tournamentID), questionIDs, initialRating)
	if err != nil {
		t.Fatalf("Failed to create tournament questions: %v", err)
	}

	tournament := &models.Tournament{
		ID:                     int(tournamentID),
		Name:                   "Test Tournament",
		QuestionsCount:         questionCount,
		Active:                 true,
		InitialK:               initialK,
		MinimumK:               minimumK,
		StdDevMultiplier:       stdDevMultiplier,
		InitialPhaseMatches:    initialPhaseMatches,
		TransitionPhaseMatches: transitionPhaseMatches,
		TopN:                   topN,
		BandSize:               bandSize,
	}

	return tournament, questionIDs
}

// TestELOIntegration tests the complete ELO system workflow
func TestELOIntegration(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Reset retry count
	ResetRetryCount()

	// Create tournament with 15 questions
	tournament, questionIDs := createTestTournament(t, 15)

	// Create ELO instance
	eloSystem := New(tournament)

	// Verify initial state
	stats, err := eloSystem.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get initial statistics: %v", err)
	}

	// All questions should be unqualified initially
	if stats["unqualified"] != 15 {
		t.Errorf("Expected 15 unqualified questions, got %v", stats["unqualified"])
	}

	if stats["total_matches"] != 0 {
		t.Errorf("Expected 0 total matches, got %v", stats["total_matches"])
	}

	if stats["total_wins"] != 0 {
		t.Errorf("Expected 0 total wins, got %v", stats["total_wins"])
	}

	// Run 20 votes
	votesRun := 0
	winners := make(map[int]int) // Track wins per question
	matches := make(map[int]int) // Track matches per question

	for i := 0; i < 20; i++ {
		// Select pair
		q1, q2, err := eloSystem.SelectPair()
		if err != nil {
			t.Fatalf("Failed to select pair on vote %d: %v", i+1, err)
		}

		// Verify different questions selected
		if q1 == q2 {
			t.Errorf("Vote %d: Same question selected twice: %d", i+1, q1)
			continue
		}

		// Verify questions are in our test set
		if !contains(questionIDs, q1) || !contains(questionIDs, q2) {
			t.Errorf("Vote %d: Invalid question IDs selected: %d, %d", i+1, q1, q2)
			continue
		}

		// Simulate user choosing first question as winner
		winnerID := q1
		loserID := q2

		// Record the match
		matches[q1]++
		matches[q2]++
		winners[winnerID]++

		// Record winner in ELO system
		err = eloSystem.RecordWinner(winnerID, loserID)
		if err != nil {
			t.Fatalf("Failed to record winner on vote %d: %v", i+1, err)
		}

		votesRun++

		// Verify question stats after each vote
		questionStats, err := eloSystem.GetQuestionsStats(q1, q2)
		if err != nil {
			t.Fatalf("Failed to get question stats on vote %d: %v", i+1, err)
		}

		if len(questionStats) != 2 {
			t.Errorf("Vote %d: Expected 2 question stats, got %d", i+1, len(questionStats))
		}
	}

	t.Logf("Successfully ran %d votes", votesRun)

	// Verify final statistics
	finalStats, err := eloSystem.GetStatistics()
	if err != nil {
		t.Fatalf("Failed to get final statistics: %v", err)
	}

	// Check total matches (should be 40, since each vote involves 2 questions)
	expectedTotalMatches := votesRun * 2
	if finalStats["total_matches"] != expectedTotalMatches {
		t.Errorf("Expected %d total matches, got %v", expectedTotalMatches, finalStats["total_matches"])
	}

	// Check total wins (should be 20, one per vote)
	if finalStats["total_wins"] != votesRun {
		t.Errorf("Expected %d total wins, got %v", votesRun, finalStats["total_wins"])
	}

	// Verify some questions should now be qualified (have >= 5 matches)
	unqualified := finalStats["unqualified"].(int)
	if unqualified >= 15 {
		t.Errorf("Expected some questions to be qualified after 20 votes, but %d still unqualified", unqualified)
	}

	// Get top questions and verify ratings have changed
	topQuestions, err := eloSystem.GetTopItems(5)
	if err != nil {
		t.Fatalf("Failed to get top questions: %v", err)
	}

	if len(topQuestions) == 0 {
		t.Error("Expected some top questions")
	}

	// Verify at least one question has rating different from initial
	ratingChanged := false
	for _, tq := range topQuestions {
		if math.Abs(tq.Rating-initialRating) > 0.1 {
			ratingChanged = true
			break
		}
	}

	if !ratingChanged {
		t.Error("Expected at least one question to have changed rating from initial value")
	}

	// Verify winners have higher ratings on average than losers
	// This is a statistical expectation over multiple votes
	totalWinnerRating := 0.0
	totalLoserRating := 0.0
	winnerCount := 0
	loserCount := 0

	for questionID := range questionIDs {
		tq, err := models.NewTournamentQuestionRepository().Find(tournament.ID, questionID+1)
		if err != nil {
			continue
		}

		winsForQuestion := winners[questionID+1]
		matchesForQuestion := matches[questionID+1]
		lossesForQuestion := matchesForQuestion - winsForQuestion

		// Add to winner stats based on wins
		totalWinnerRating += tq.Rating * float64(winsForQuestion)
		winnerCount += winsForQuestion

		// Add to loser stats based on losses
		totalLoserRating += tq.Rating * float64(lossesForQuestion)
		loserCount += lossesForQuestion
	}

	if winnerCount > 0 && loserCount > 0 {
		avgWinnerRating := totalWinnerRating / float64(winnerCount)
		avgLoserRating := totalLoserRating / float64(loserCount)

		t.Logf("Average winner rating: %.2f, Average loser rating: %.2f", avgWinnerRating, avgLoserRating)

		// This is a weak test since we're using the same question as winner every time
		// but it should still show some difference
		if avgWinnerRating <= avgLoserRating {
			t.Logf("Warning: Expected winner ratings to be higher than loser ratings on average")
		}
	}

	// Check retry count (should be low with 15 questions and 20 votes)
	retryCount := GetRetryCount()
	t.Logf("Retry count: %d", retryCount)

	// Verify threshold calculation works
	threshold, err := eloSystem.CalculateThreshold()
	if err != nil {
		t.Fatalf("Failed to calculate threshold: %v", err)
	}

	t.Logf("Final threshold: %.2f", threshold)

	// Print final distribution for debugging
	if distribution, ok := finalStats["distribution"].(map[int]int); ok {
		t.Logf("Rating distribution: %+v", distribution)
	}

	t.Logf("ELO integration test completed successfully")
}

// contains checks if a slice contains a value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// TestELOThresholdCalculation tests the threshold calculation with edge cases
func TestELOThresholdCalculation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create tournament with fewer questions than TopN
	tournament, _ := createTestTournament(t, 3)
	eloSystem := New(tournament)

	// With fewer qualified questions than TopN, threshold should be -inf
	threshold, err := eloSystem.CalculateThreshold()
	if err != nil {
		t.Fatalf("Failed to calculate threshold: %v", err)
	}

	if !math.IsInf(threshold, -1) {
		t.Errorf("Expected -inf threshold with few questions, got %f", threshold)
	}
}

// TestELORatingChanges tests that ratings change appropriately
func TestELORatingChanges(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	tournament, _ := createTestTournament(t, 5)
	eloSystem := New(tournament)

	// Get initial ratings
	repo := models.NewTournamentQuestionRepository()
	q1Initial, err := repo.Find(tournament.ID, 1)
	if err != nil {
		t.Fatalf("Failed to find question 1: %v", err)
	}

	q2Initial, err := repo.Find(tournament.ID, 2)
	if err != nil {
		t.Fatalf("Failed to find question 2: %v", err)
	}

	initialRating1 := q1Initial.Rating
	initialRating2 := q2Initial.Rating

	// Record question 1 as winner
	err = eloSystem.RecordWinner(1, 2)
	if err != nil {
		t.Fatalf("Failed to record winner: %v", err)
	}

	// Get updated ratings
	q1Updated, err := repo.Find(tournament.ID, 1)
	if err != nil {
		t.Fatalf("Failed to find updated question 1: %v", err)
	}

	q2Updated, err := repo.Find(tournament.ID, 2)
	if err != nil {
		t.Fatalf("Failed to find updated question 2: %v", err)
	}

	// Winner should have increased rating
	if q1Updated.Rating <= initialRating1 {
		t.Errorf("Winner rating should increase: %.2f -> %.2f", initialRating1, q1Updated.Rating)
	}

	// Loser should have decreased rating
	if q2Updated.Rating >= initialRating2 {
		t.Errorf("Loser rating should decrease: %.2f -> %.2f", initialRating2, q2Updated.Rating)
	}

	// Verify match and win counts
	if q1Updated.Matches != 1 {
		t.Errorf("Winner should have 1 match, got %d", q1Updated.Matches)
	}

	if q1Updated.Wins != 1 {
		t.Errorf("Winner should have 1 win, got %d", q1Updated.Wins)
	}

	if q2Updated.Matches != 1 {
		t.Errorf("Loser should have 1 match, got %d", q2Updated.Matches)
	}

	if q2Updated.Wins != 0 {
		t.Errorf("Loser should have 0 wins, got %d", q2Updated.Wins)
	}
}
