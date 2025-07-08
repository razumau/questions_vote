package models

import "time"

// Question represents a question
type Question struct {
	ID             int      `json:"id"`
	GotQuestionsID int      `json:"gotquestions_id,omitempty"`
	Question       string   `json:"question"`
	Answer         string   `json:"answer"`
	AcceptedAnswer string   `json:"accepted_answer,omitempty"`
	Comment        string   `json:"comment"`
	Source         string   `json:"source"`
	HandoutStr     string   `json:"handout_str,omitempty"`
	HandoutImg     string   `json:"handout_img,omitempty"`
	AuthorID       *int     `json:"author_id,omitempty"`
	PackageID      *int     `json:"package_id,omitempty"`
	Difficulty     *float64 `json:"difficulty,omitempty"`
	IsIncorrect    *bool    `json:"is_incorrect,omitempty"`
}

// Vote represents a user's vote between two questions
type Vote struct {
	ID           int       `json:"id"`
	UserID       int64     `json:"user_id"`
	Question1ID  int       `json:"question1_id"`
	Question2ID  int       `json:"question2_id"`
	TournamentID int       `json:"tournament_id"`
	SelectedID   *int      `json:"selected_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// Tournament represents a tournament with questions
type Tournament struct {
	ID                     int     `json:"id"`
	Name                   string  `json:"name"`
	QuestionsCount         int     `json:"questions_count"`
	Active                 bool    `json:"active"`
	InitialK               float64 `json:"initial_k"`
	MinimumK               float64 `json:"minimum_k"`
	StdDevMultiplier       float64 `json:"std_dev_multiplier"`
	InitialPhaseMatches    int     `json:"initial_phase_matches"`
	TransitionPhaseMatches int     `json:"transition_phase_matches"`
	TopN                   int     `json:"top_n"`
	BandSize               int     `json:"band_size"`
}

// QuestionStats represents statistics for a question
type QuestionStats struct {
	Wins    int `json:"wins"`
	Matches int `json:"matches"`
}

// VoteChoice represents the choice made in a vote callback
type VoteChoice struct {
	Question1ID int
	Question2ID int
	Choice      int // 0 = skip, 1 = first question, 2 = second question
}
