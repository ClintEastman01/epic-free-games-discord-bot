package service

import (
	"fmt"
	"log"
	"time"

	"free-games-scrape/internal/database"
	"free-games-scrape/internal/models"
	"free-games-scrape/internal/scraper"
)

// GameService handles game-related business logic
type GameService struct {
	db      *database.Database
	scraper *scraper.EpicScraper
}

// NewGameService creates a new game service
func NewGameService(db *database.Database, scraper *scraper.EpicScraper) *GameService {
	return &GameService{
		db:      db,
		scraper: scraper,
	}
}

// RefreshGames scrapes new games and updates the database
func (gs *GameService) RefreshGames() error {
	log.Println("Starting game refresh...")
	
	// Scrape games from Epic Games Store
	scrapedGames, err := gs.scraper.ScrapeGames()
	if err != nil {
		return fmt.Errorf("failed to scrape games: %w", err)
	}

	if len(scrapedGames) == 0 {
		log.Println("No games found during scraping")
		return nil
	}

	// Save games to database
	if err := gs.db.SaveGames(scrapedGames); err != nil {
		return fmt.Errorf("failed to save games to database: %w", err)
	}

	// Cleanup old games
	if err := gs.db.CleanupOldGames(); err != nil {
		log.Printf("Warning: failed to cleanup old games: %v", err)
	}

	log.Printf("Successfully refreshed %d games", len(scrapedGames))
	return nil
}

// GetActiveGames returns all currently active games from the database
func (gs *GameService) GetActiveGames() (*models.GameCollection, error) {
	games, err := gs.db.GetActiveGames()
	if err != nil {
		return nil, fmt.Errorf("failed to get active games: %w", err)
	}

	return models.NewGameCollection(games), nil
}

// GetNewGamesSince returns games that are new since the specified time
func (gs *GameService) GetNewGamesSince(since time.Time) (*models.GameCollection, error) {
	games, err := gs.db.GetNewGames(since)
	if err != nil {
		return nil, fmt.Errorf("failed to get new games: %w", err)
	}

	return models.NewGameCollection(games), nil
}

// GetGameByTitle retrieves a specific game by title
func (gs *GameService) GetGameByTitle(title string) (*models.Game, error) {
	return gs.db.GetGameByTitle(title)
}

// ShouldRefresh determines if games should be refreshed based on cache age
func (gs *GameService) ShouldRefresh(maxAge time.Duration) (bool, error) {
	// For now, we'll refresh based on time intervals
	// In a more sophisticated implementation, you could track last refresh time in the database
	return true, nil
}