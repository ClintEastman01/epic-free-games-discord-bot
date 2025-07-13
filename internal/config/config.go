package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Discord  DiscordConfig
	Scraper  ScraperConfig
	Database DatabaseConfig
}

// DiscordConfig holds Discord-related configuration
type DiscordConfig struct {
	BotToken  string
	ChannelID string
}

// ScraperConfig holds scraper-related configuration
type ScraperConfig struct {
	ChromePath string
	UserAgent  string
	Timeout    int // in seconds
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist, just log a warning
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	config := &Config{
		Discord: DiscordConfig{
			BotToken:  os.Getenv("DISCORD_BOT_TOKEN"),
			ChannelID: os.Getenv("YOUR_CHANNEL_ID"),
		},
		Scraper: ScraperConfig{
			ChromePath: getDefaultChromePath(),
			UserAgent:  "Mozilla/5.0 (compatible; EpicGamesBotScraper/1.0)",
			Timeout:    90,
		},
		Database: DatabaseConfig{
			Path: getEnvOrDefault("DATABASE_PATH", "games.db"),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	if c.Discord.BotToken == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN environment variable is required")
	}
	if c.Discord.ChannelID == "" {
		return fmt.Errorf("YOUR_CHANNEL_ID environment variable is required")
	}
	return nil
}

// getDefaultChromePath returns the default Chrome/Chromium path based on OS
func getDefaultChromePath() string {
	switch runtime.GOOS {
	case "darwin": // macOS
		paths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	case "linux":
		paths := []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		return "/usr/bin/google-chrome"
	case "windows":
		paths := []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		return "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
	default:
		return ""
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}