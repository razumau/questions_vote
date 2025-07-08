package importer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"questions-vote/internal/db"
	"questions-vote/internal/models"
)

// PackageParser fetches and stores questions from a specific package
type PackageParser struct {
	PackageID int
	URL       string
	Rewrite   bool

	questionRepo *models.QuestionRepository
}

// NewPackageParser creates a new package parser
func NewPackageParser(packageID int, rewrite bool) *PackageParser {
	return &PackageParser{
		PackageID:    packageID,
		URL:          fmt.Sprintf("https://gotquestions.online/pack/%d", packageID),
		Rewrite:      rewrite,
		questionRepo: models.NewQuestionRepository(),
	}
}

// ImportPackage imports all questions from the package
func (pp *PackageParser) ImportPackage() error {
	log.Printf("Importing package %d from %s", pp.PackageID, pp.URL)

	if pp.Rewrite {
		err := pp.deleteOldEntries()
		if err != nil {
			return fmt.Errorf("failed to delete old entries: %w", err)
		}
	}

	nextJsData, err := ExtractNextJsDataFromURL(pp.URL)
	if err != nil {
		return fmt.Errorf("failed to extract data from URL %s: %w", pp.URL, err)
	}

	packKeyValue, err := FindKeyInData(nextJsData, "pack")
	if err != nil {
		return fmt.Errorf("could not find pack key at URL %s: %w", pp.URL, err)
	}

	pack, ok := packKeyValue.(map[string]any)
	if !ok {
		return fmt.Errorf("pack was not a dict at URL %s", pp.URL)
	}

	questions, err := pp.extractQuestions(pack)
	if err != nil {
		return fmt.Errorf("failed to extract questions: %w", err)
	}

	log.Printf("Found %d questions in package %d", len(questions), pp.PackageID)

	for i, questionDict := range questions {
		question, err := models.BuildQuestionFromDict(questionDict, pp.PackageID)
		if err != nil {
			log.Printf("Failed to build question %d: %v", i, err)
			continue
		}

		err = pp.insertQuestion(question)
		if err != nil {
			log.Printf("Failed to insert question %d: %v", i, err)
			continue
		}
	}

	log.Printf("Completed importing package %d", pp.PackageID)
	return nil
}

// extractQuestions extracts questions from the props data
func (pp *PackageParser) extractQuestions(pack map[string]interface{}) ([]map[string]interface{}, error) {
	var allQuestions []map[string]interface{}

	toursInterface, ok := pack["tours"]
	if !ok {
		return nil, fmt.Errorf("no tours found in pack")
	}

	tours, ok := toursInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tours structure")
	}

	for _, tourInterface := range tours {
		tour, ok := tourInterface.(map[string]interface{})
		if !ok {
			continue
		}

		questionsInterface, ok := tour["questions"]
		if !ok {
			continue
		}

		questions, ok := questionsInterface.([]interface{})
		if !ok {
			continue
		}

		for _, questionInterface := range questions {
			questionDict, ok := questionInterface.(map[string]interface{})
			if ok {
				allQuestions = append(allQuestions, questionDict)
			}
		}
	}

	// Try to get from pack.tours structure
	// if packInterface, exists := pagePropsPack["pack"]; exists {

	// } else if tourInterface, exists := pagePropsPack["tour"]; exists {
	// 	// Try to get from direct tour structure
	// 	tour, ok := tourInterface.(map[string]interface{})
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid tour structure")
	// 	}

	// 	questionsInterface, ok := tour["questions"]
	// 	if !ok {
	// 		return nil, fmt.Errorf("no questions found in tour")
	// 	}

	// 	questions, ok := questionsInterface.([]interface{})
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid questions structure")
	// 	}

	// 	for _, questionInterface := range questions {
	// 		questionDict, ok := questionInterface.(map[string]interface{})
	// 		if ok {
	// 			allQuestions = append(allQuestions, questionDict)
	// 		}
	// 	}
	// } else {
	// 	return nil, fmt.Errorf("no pack or tour data found")
	// }

	return allQuestions, nil
}

// deleteOldEntries removes existing questions for this package
func (pp *PackageParser) deleteOldEntries() error {
	database := db.GetDB()

	// Delete images first (foreign key constraint)
	_, err := database.Exec(`
		DELETE FROM images 
		WHERE question_id IN (SELECT id FROM questions WHERE package_id = ?)
	`, pp.PackageID)
	if err != nil {
		return fmt.Errorf("failed to delete old images: %w", err)
	}

	// Delete questions
	_, err = database.Exec("DELETE FROM questions WHERE package_id = ?", pp.PackageID)
	if err != nil {
		return fmt.Errorf("failed to delete old questions: %w", err)
	}

	return nil
}

// insertQuestion inserts a question and its image if present
func (pp *PackageParser) insertQuestion(question *models.Question) error {
	database := db.GetDB()

	// Insert question
	result, err := database.Exec(`
		INSERT INTO questions (
			gotquestions_id, question, answer, accepted_answer, comment, 
			handout_str, source, author_id, package_id, difficulty, is_incorrect
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		question.GotQuestionsID, question.Question, question.Answer, question.AcceptedAnswer,
		question.Comment, question.HandoutStr, question.Source, question.AuthorID,
		question.PackageID, question.Difficulty, question.IsIncorrect,
	)
	if err != nil {
		return fmt.Errorf("failed to insert question: %w", err)
	}

	questionID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get question ID: %w", err)
	}

	if question.HandoutImg != "" {
		err = pp.insertImage(int(questionID), question.HandoutImg)
		if err != nil {
			log.Printf("Failed to insert image for question %d: %v", questionID, err)
		}
	}

	return nil
}

// insertImage downloads and inserts an image for a question
func (pp *PackageParser) insertImage(questionID int, handoutImg string) error {
	imageURL := fmt.Sprintf("https://gotquestions.online/%s", handoutImg)

	response, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image from %s: %w", imageURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d when downloading image from %s", response.StatusCode, imageURL)
	}

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	mimeType := response.Header.Get("Content-Type")

	database := db.GetDB()
	_, err = database.Exec(`
		INSERT INTO images (question_id, image_url, data, mime_type) 
		VALUES (?, ?, ?, ?)
	`, questionID, handoutImg, imageData, mimeType)
	if err != nil {
		return fmt.Errorf("failed to insert image data: %w", err)
	}

	return nil
}
