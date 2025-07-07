package importer

import (
	"fmt"
	"log"
	"math/rand"
	"questions-vote/internal/models"
	"time"
)

// PackageLister fetches and stores package information from gotquestions.online
type PackageLister struct {
	FirstPage int
	LastPage  int
	repo      *models.PackageRepository
}

// NewPackageLister creates a new package lister
func NewPackageLister(firstPage, lastPage int) *PackageLister {
	return &PackageLister{
		FirstPage: firstPage,
		LastPage:  lastPage,
		repo:      models.NewPackageRepository(),
	}
}

// Run processes all pages from first to last
func (pl *PackageLister) Run() error {
	log.Printf("Starting package listing from page %d to %d", pl.FirstPage, pl.LastPage)

	for page := pl.FirstPage; page <= pl.LastPage; page++ {
		if page%10 == 1 {
			log.Printf("Processing page %d", page)
		}

		err := pl.CreatePackagesFromPage(page)
		if err != nil {
			log.Printf("Error processing page %d: %v", page, err)
			// Continue with next page instead of failing completely
			continue
		}

		// Sleep to avoid rate limiting
		SleepAround(1, 0.5)
	}

	log.Printf("Completed package listing")
	return nil
}

// CreatePackagesFromPage fetches packages from a specific page
func (pl *PackageLister) CreatePackagesFromPage(page int) error {
	url := fmt.Sprintf("https://gotquestions.online/?page=%d", page)

	nextJsData, err := ExtractNextJsDataFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to extract data from page %d: %w", page, err)
	}

	packsKeyValue, err := FindKeyInData(nextJsData, "packs")
	if err != nil {
		return fmt.Errorf("failed to find packs in page %d: %w", page, err)
	}

	packs, ok := packsKeyValue.([]any)
	if !ok {
		return fmt.Errorf("packs was not a list in page %d: %w", page, err)
	}

	log.Printf("Found %d packs", len(packs))
	for _, packInterface := range packs {
		packDict, ok := packInterface.(map[string]any)
		if !ok {
			log.Printf("Skipping invalid package structure on page %d", page)
			continue
		}

		pkg, err := models.BuildPackageFromDict(packDict)
		if err != nil {
			log.Printf("Failed to build package from data on page %d: %v", page, err)
			continue
		}

		err = pl.repo.Insert(pkg)
		if err != nil {
			log.Printf("Failed to insert package %d: %v", pkg.GotQuestionsID, err)
			continue
		}
		log.Printf("Inserted package %d", pkg.GotQuestionsID)
	}

	return nil
}

// SleepAround sleeps for a random duration around the specified seconds
func SleepAround(seconds float64, deviation float64) {
	min := seconds - deviation
	max := seconds + deviation
	if min < 0 {
		min = 0
	}

	duration := min + rand.Float64()*(max-min)
	time.Sleep(time.Duration(duration * float64(time.Second)))
}
