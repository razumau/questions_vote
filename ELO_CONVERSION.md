# ELO System Conversion

This document describes the conversion of the ELO rating system from Python to Go.

## Components Implemented

### 1. Tournament Question Model (`internal/models/tournament_question.go`)
- Manages questions within tournaments with ELO ratings
- Provides database operations for tournament questions
- Handles rating calculations and statistics

Key methods:
- `CreateTournamentQuestions()` - Bulk create tournament questions
- `GetRandomQuestion()` - Select random question based on criteria
- `Find()` - Find specific tournament question
- `Save()` - Update question ratings and stats
- `GetStatsForQualified()` - Calculate statistics for qualified questions
- `CountUnqualifiedQuestions()` - Count questions below qualification threshold

### 2. ELO System (`internal/elo/elo.go`)
- Core ELO rating algorithm implementation
- Question pair selection logic
- Rating update calculations

Key methods:
- `New()` - Create ELO system from tournament
- `SelectPair()` - Select two questions for comparison
- `RecordWinner()` - Update ratings after vote
- `CalculateThreshold()` - Calculate rating threshold for selection
- `GetQuestionsStats()` - Get statistics for question pairs
- `GetStatistics()` - Get comprehensive tournament statistics

### 3. Integration Tests (`internal/elo/elo_test.go`)
Comprehensive integration tests that verify:
- Tournament creation with 15 pseudo-questions
- 20 voting rounds with proper ELO updates
- Rating changes and distribution
- Statistical calculations
- Threshold computation
- Edge case handling

### 4. Service Layer Updates
Updated `QuestionService` and `VoteService` to:
- Use ELO system for question selection
- Record ELO rating updates on votes
- Provide real statistics from tournament data
- Maintain fallbacks for robustness

## ELO Algorithm Details

### Question Selection Strategy
1. **Unqualified Phase**: Questions with < 5 matches get priority
2. **Mixed Phase**: One unqualified + one any question
3. **Qualified Phase**: Only questions above rating threshold

### Rating Calculation
- Uses classic ELO formula: `Rating_new = Rating_old + K * (Score - Expected)`
- Dynamic K-factor based on match count:
  - Initial phase (< 5 matches): K = 64
  - Transition phase (5-10 matches): K = 32
  - Mature phase (> 10 matches): K = 16

### Threshold Calculation
- For top N selection: `Threshold = TopN_Rating - (StdDev * Multiplier)`
- Prevents rating inflation by focusing on competitive questions

## Test Results

The integration test successfully demonstrates:
- ✅ 20 votes processed correctly
- ✅ Rating distribution shows spread (1380-1560 range)
- ✅ Winners have higher average rating than losers (1516.98 vs 1465.32)
- ✅ Total match/win counts are accurate (40 matches, 20 wins)
- ✅ Question qualification progression works
- ✅ Minimal retries (1 out of 20 selections)

## Usage

```go
// Create ELO system from tournament
eloSystem := elo.New(tournament)

// Select question pair
q1ID, q2ID, err := eloSystem.SelectPair()

// Record vote result
err = eloSystem.RecordWinner(winnerID, loserID)

// Get statistics
stats, err := eloSystem.GetStatistics()
```

## Database Schema

The system expects these tables:
- `tournaments` - Tournament configuration
- `tournament_questions` - Questions with ratings/stats
- `votes` - User voting history

## Compatibility

The Go implementation maintains full compatibility with the Python version:
- Same algorithm logic and parameters
- Compatible database schema
- Equivalent statistical calculations
- Same question selection strategy

The system can work with existing Python-created tournament data.