package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"questions-vote/internal/db"
	"questions-vote/internal/importer"
	"questions-vote/internal/models"
)

func main() {
	var (
		command = flag.String("command", "", "Command to run: list-packages, import-package, import-year")
		firstPage = flag.Int("first-page", 1, "First page to process for list-packages")
		lastPage = flag.Int("last-page", 337, "Last page to process for list-packages")
		packageID = flag.Int("package-id", 0, "Package ID to import for import-package")
		year = flag.Int("year", 0, "Year to import for import-year")
		rewrite = flag.Bool("rewrite", true, "Rewrite existing questions when importing packages")
	)
	flag.Parse()

	if *command == "" {
		printUsage()
		os.Exit(1)
	}

	// Initialize database
	err := db.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	switch *command {
	case "list-packages":
		err = runListPackages(*firstPage, *lastPage)
	case "import-package":
		if *packageID == 0 {
			log.Fatal("Package ID is required for import-package command")
		}
		err = runImportPackage(*packageID, *rewrite)
	case "import-year":
		if *year == 0 {
			log.Fatal("Year is required for import-year command")
		}
		err = runImportYear(*year, *rewrite)
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
	fmt.Println("Questions Vote Importer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  importer -command=list-packages [-first-page=1] [-last-page=337]")
	fmt.Println("    Updates the list of packages in the database")
	fmt.Println()
	fmt.Println("  importer -command=import-package -package-id=ID [-rewrite=true]")
	fmt.Println("    Fetches questions for a specific package ID")
	fmt.Println()
	fmt.Println("  importer -command=import-year -year=YEAR [-rewrite=true]")
	fmt.Println("    Fetches questions from all packages for a specific year")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  importer -command=list-packages")
	fmt.Println("  importer -command=import-package -package-id=5220")
	fmt.Println("  importer -command=import-year -year=2022")
}

func runListPackages(firstPage, lastPage int) error {
	log.Printf("Starting package listing from page %d to %d", firstPage, lastPage)
	
	lister := importer.NewPackageLister(firstPage, lastPage)
	err := lister.Run()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}
	
	log.Println("Package listing completed successfully")
	return nil
}

func runImportPackage(packageID int, rewrite bool) error {
	log.Printf("Starting import of package %d (rewrite: %v)", packageID, rewrite)
	
	parser := importer.NewPackageParser(packageID, rewrite)
	err := parser.ImportPackage()
	if err != nil {
		return fmt.Errorf("failed to import package %d: %w", packageID, err)
	}
	
	log.Printf("Package %d imported successfully", packageID)
	return nil
}

func runImportYear(year int, rewrite bool) error {
	log.Printf("Starting import of all packages for year %d (rewrite: %v)", year, rewrite)
	
	// Get all packages for the year
	repo := models.NewPackageRepository()
	packages, err := repo.GetPackagesByYear(year)
	if err != nil {
		return fmt.Errorf("failed to get packages for year %d: %w", year, err)
	}
	
	if len(packages) == 0 {
		log.Printf("No packages found for year %d", year)
		return nil
	}
	
	log.Printf("Found %d packages for year %d", len(packages), year)
	
	// Check if packages already have questions (if not rewriting)
	questionRepo := models.NewQuestionRepository()
	
	successCount := 0
	errorCount := 0
	
	for i, pkg := range packages {
		log.Printf("Processing package %d/%d: %d (%s)", i+1, len(packages), pkg.GotQuestionsID, pkg.Title)
		
		// Check if package already has questions
		if !rewrite {
			hasQuestions, err := questionRepo.HasQuestionsFromPackage(pkg.GotQuestionsID)
			if err != nil {
				log.Printf("Failed to check if package %d has questions: %v", pkg.GotQuestionsID, err)
				continue
			}
			if hasQuestions {
				log.Printf("Skipping package %d (already has questions)", pkg.GotQuestionsID)
				continue
			}
		}
		
		// Import the package
		parser := importer.NewPackageParser(pkg.GotQuestionsID, rewrite)
		err = parser.ImportPackage()
		if err != nil {
			log.Printf("Failed to import package %d: %v", pkg.GotQuestionsID, err)
			errorCount++
			continue
		}
		
		successCount++
		
		// Sleep between packages to avoid rate limiting
		importer.SleepAround(1.0, 0.7)
	}
	
	log.Printf("Year %d import completed: %d successful, %d errors", year, successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors out of %d packages", errorCount, len(packages))
	}
	
	return nil
}