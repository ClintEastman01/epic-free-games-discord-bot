package app

import (
	"context"
	"free-games-scrape/internal/bot"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/database"
	"free-games-scrape/internal/logger"
	"free-games-scrape/internal/metrics"
	"free-games-scrape/internal/models"
	"free-games-scrape/internal/ratelimit"
	"free-games-scrape/internal/scraper"
	"free-games-scrape/internal/security"
	"free-games-scrape/internal/service"
	"free-games-scrape/internal/web"
	"log"
	"os"
	"os/signal"
	"time"
)

// App represents the main application
type App struct {
	config      *config.Config
	discordBot  *bot.DiscordBot
	gameService *service.GameService
	db          *database.Database
	webServer   *web.WebServer
	logger      *logger.Logger
	metrics     *metrics.Metrics
	rateLimiter *ratelimit.DiscordRateLimiter
	validator   *security.Validator
	lastCheck   time.Time
	ctx         context.Context
	cancel      context.CancelFunc
}

// New creates a new application instance with enhanced features
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Initialize logger
	appLogger := logger.New(logger.LogLevel(cfg.App.LogLevel), cfg.App.Environment)
	appLogger.Info("Starting Free Games Bot v2.0")

	// Validate Discord token
	validator := security.NewValidator()
	if err := validator.ValidateDiscordToken(cfg.Discord.Token); err != nil {
		return nil, err
	}

	// Initialize metrics
	appMetrics := metrics.New()

	// Initialize rate limiter
	rateLimiter := ratelimit.NewDiscordRateLimiter()

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	// Initialize Epic Games scraper
	epicScraper := scraper.NewEpicScraper(&cfg.Scraper)

	// Initialize game service
	gameService := service.NewGameService(db, epicScraper)

	// Initialize Discord bot with game service and database
	discordBot, err := bot.NewDiscordBot(&cfg.Discord, gameService, db)
	if err != nil {
		return nil, err
	}

	// Initialize web server for documentation
	webServer := web.NewWebServer(cfg.Web.Port, gameService, db)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		config:      cfg,
		discordBot:  discordBot,
		gameService: gameService,
		db:          db,
		webServer:   webServer,
		logger:      appLogger,
		metrics:     appMetrics,
		rateLimiter: rateLimiter,
		validator:   validator,
		lastCheck:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start web server in a goroutine
	go func() {
		log.Println("Starting web server for documentation...")
		if err := a.webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	// Start Discord bot
	if err := a.discordBot.Start(); err != nil {
		return err
	}
	defer a.discordBot.Stop()
	defer a.db.Close()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Run initial scraping immediately on startup
	log.Println("Running initial game check...")
	if err := a.performGameCheck(); err != nil {
		log.Printf("Initial scraping failed: %v", err)
		a.discordBot.SendErrorMessage("Failed to perform initial game check. Will retry in 24 hours.")
	}

	// Ticker for periodic scraping (every 6 hours for more frequent updates)
	ticker := time.NewTicker(6 * time.Hour)
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
				a.discordBot.SendErrorMessage("Failed to check for free games. Will retry in 6 hours.")
			}
		}
	}
}

// performGameCheck scrapes games and sends updates for new games only
func (a *App) performGameCheck() error {
	// Scrape games from Epic Games Store
	scrapedGames, err := a.gameService.ScrapeGames()
	if err != nil {
		return err
	}

	if len(scrapedGames) == 0 {
		log.Println("No games found during scraping")
		return nil
	}

	// Get current games from database to compare
	currentGames, err := a.gameService.GetActiveGames()
	if err != nil {
		return err
	}

	// Find truly new games by comparing scraped games with database
	newGames := a.findNewGames(scrapedGames, currentGames)

	// Save all scraped games to database (updates existing, adds new)
	if err := a.gameService.SaveGames(scrapedGames); err != nil {
		return err
	}

	// Send updates to Discord only for new games
	if len(newGames.FreeNow) > 0 || len(newGames.ComingSoon) > 0 {
		if err := a.discordBot.SendGameUpdates(newGames); err != nil {
			return err
		}
		log.Printf("Sent updates for %d new Free Now games and %d new Coming Soon games",
			len(newGames.FreeNow), len(newGames.ComingSoon))
	} else {
		log.Println("No new games found since last check")
	}

	// Update last check time
	a.lastCheck = time.Now()

	return nil
}

// findNewGames compares scraped games with current database games to find truly new ones
func (a *App) findNewGames(scrapedGames []models.Game, currentGames *models.GameCollection) *models.GameCollection {
	// Create a map of existing games with their free-to dates for quick lookup
	// Key format: "GameTitle|FreeTo" to handle cases where the same game becomes free again
	existingGames := make(map[string]bool)
	
	// Add all current games to the map
	for _, game := range currentGames.FreeNow {
		key := game.Title + "|" + game.FreeTo
		existingGames[key] = true
	}
	for _, game := range currentGames.ComingSoon {
		key := game.Title + "|" + game.FreeTo
		existingGames[key] = true
	}

	// Find games that are in scraped but not in existing with the same free-to date
	var newGames []models.Game
	for _, game := range scrapedGames {
		key := game.Title + "|" + game.FreeTo
		if !existingGames[key] {
			newGames = append(newGames, game)
			log.Printf("Found new game: %s (Status: %s, Free until: %s)", 
				game.Title, game.Status, game.FreeTo)
		}
	}

	return models.NewGameCollection(newGames)
}

