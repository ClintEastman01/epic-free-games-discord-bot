package app

import (
	"log"
	"os"
	"os/signal"
	"time"

	"free-games-scrape/internal/bot"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/models"
	"free-games-scrape/internal/scraper"
)

// App represents the main application
type App struct {
	config      *config.Config
	discordBot  *bot.DiscordBot
	epicScraper *scraper.EpicScraper
}

// New creates a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Initialize Discord bot
	discordBot, err := bot.NewDiscordBot(&cfg.Discord)
	if err != nil {
		return nil, err
	}

	// Initialize Epic Games scraper
	epicScraper := scraper.NewEpicScraper(&cfg.Scraper)

	return &App{
		config:      cfg,
		discordBot:  discordBot,
		epicScraper: epicScraper,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start Discord bot
	if err := a.discordBot.Start(); err != nil {
		return err
	}
	defer a.discordBot.Stop()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run initial scraping immediately on startup
	log.Println("Running initial game check...")
	if err := a.performGameCheck(); err != nil {
		log.Printf("Initial scraping failed: %v", err)
		a.discordBot.SendErrorMessage("Failed to perform initial game check. Will retry in 24 hours.")
	}

	// Ticker for periodic scraping (every 24 hours after initial run)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	log.Println("Bot is now running. Press Ctrl+C to stop.")

	for {
		select {
		case <-stop:
			log.Println("Received shutdown signal")
			return nil
		case <-ticker.C:
			log.Println("Performing scheduled game check...")
			if err := a.performGameCheck(); err != nil {
				log.Printf("Scheduled scraping failed: %v", err)
				a.discordBot.SendErrorMessage("Failed to check for free games. Will retry in 24 hours.")
			}
		}
	}
}

// performGameCheck scrapes games and sends updates
func (a *App) performGameCheck() error {
	// Scrape games from Epic Games Store
	games, err := a.epicScraper.ScrapeGames()
	if err != nil {
		return err
	}

	if len(games) == 0 {
		log.Println("No games found during scraping")
		return nil
	}

	// Create game collection
	gameCollection := models.NewGameCollection(games)

	// Send updates to Discord
	if err := a.discordBot.SendGameUpdates(gameCollection); err != nil {
		return err
	}

	// Check if we should continue running
	if !gameCollection.HasActiveFreeGames() {
		log.Println("No active Free Now games remaining. Sending notification.")
		a.discordBot.SendSimpleMessage("ℹ️ No active free games remaining. Bot will continue monitoring for new free games.")
	}

	log.Printf("Successfully processed %d games (%d Free Now, %d Coming Soon)", 
		len(games), len(gameCollection.FreeNow), len(gameCollection.ComingSoon))

	return nil
}