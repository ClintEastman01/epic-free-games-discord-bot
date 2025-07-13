package scraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/models"
)

// EpicScraper handles scraping Epic Games Store for free games
type EpicScraper struct {
	config *config.ScraperConfig
}

// NewEpicScraper creates a new Epic Games scraper
func NewEpicScraper(cfg *config.ScraperConfig) *EpicScraper {
	return &EpicScraper{
		config: cfg,
	}
}

// ScrapeGames scrapes free games from Epic Games Store
func (s *EpicScraper) ScrapeGames() ([]models.Game, error) {
	// Create context with Chrome executable path
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.ExecPath(s.config.ChromePath),
		chromedp.UserAgent(s.config.UserAgent),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, time.Duration(s.config.Timeout)*time.Second)
	defer cancel()

	var games []models.Game

	// Attempt to scrape with retries
	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("Scraping attempt %d/3", attempt)
		
		err := chromedp.Run(ctx,
			chromedp.Navigate("https://store.epicgames.com/en-US/free-games"),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Sleep(2*time.Second), // Wait for dynamic content to load
			chromedp.Evaluate(s.getScrapingScript(), &games),
		)
		
		if err == nil && len(games) > 0 {
			log.Printf("Successfully scraped %d games", len(games))
			return games, nil
		}
		
		log.Printf("Attempt %d failed: %v. Retrying...", attempt, err)
		if attempt < 3 {
			time.Sleep(5 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to scrape data after 3 attempts")
}

// getScrapingScript returns the JavaScript code for scraping game data
func (s *EpicScraper) getScrapingScript() string {
	return `
		(() => {
			const games = [];
			const containers = document.querySelectorAll('[data-component="FreeOfferCard"]');
			
			if (containers.length === 0) {
				console.log('No FreeOfferCard containers found');
				return games;
			}
			
			containers.forEach((container, index) => {
				try {
					const game = {};
					
					// Extract title
					const titleElement = container.querySelector('.css-1p5cyzj-ROOT h6, h6, [data-testid="offer-title"]');
					game.title = titleElement?.textContent?.trim() || '';
					
					// Extract image URL
					const imageElement = container.querySelector('img[data-image], img[src]');
					game.image_url = imageElement?.getAttribute('data-image') || imageElement?.getAttribute('src') || '';
					
					// Extract status
					const statusElement = container.querySelector('.css-82y1uz span, .css-gyjcm9 span, [data-testid="offer-status"]');
					game.status = statusElement?.textContent?.trim() || '';
					
					// Extract period information
					const periodElement = container.querySelector('.css-1p5cyzj-ROOT p span, [data-testid="offer-period"]');
					const period = periodElement?.textContent?.trim() || '';
					
					if (period.includes('Free Now')) {
						const parts = period.split(' - ');
						game.free_to = parts.length > 1 ? parts[1].split(' at ')[0].trim() : '';
					} else if (period.includes('Free')) {
						const parts = period.split(' - ');
						if (parts.length > 1) {
							game.free_from = parts[0].replace('Free', '').trim();
							game.free_to = parts[1].trim();
						}
					}
					
					// Only add games with valid titles
					if (game.title) {
						games.push(game);
						console.log('Found game:', game.title, 'Status:', game.status);
					}
				} catch (error) {
					console.error('Error processing game container', index, ':', error);
				}
			});
			
			console.log('Total games found:', games.length);
			return games;
		})()
	`
}