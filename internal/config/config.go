package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Discord  DiscordConfig
	Scraper  ScraperConfig
	Database DatabaseConfig
	Web      WebConfig
	App      AppConfig
}

// DiscordConfig holds Discord-specific configuration
type DiscordConfig struct {
	Token           string
	ClientID        string
	ChannelID       string
	MaxRetries      int
	RetryDelay      time.Duration
	CommandTimeout  time.Duration
	RateLimitBuffer time.Duration
}

// ScraperConfig holds scraper-specific configuration
type ScraperConfig struct {
	ChromePath   string
	UserAgent    string
	Timeout      time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	RequestDelay time.Duration
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Path              string
	MaxConnections    int
	ConnectionTimeout time.Duration
	QueryTimeout      time.Duration
}

// WebConfig holds web server configuration
type WebConfig struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Environment     string
	LogLevel        string
	RefreshInterval time.Duration
	GracefulTimeout time.Duration
}

// Load loads configuration from environment variables with validation
func Load() (*Config, error) {
	// Discord configuration
	token := strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN"))
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN environment variable is required")
	}

	clientID := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_ID"))
	if clientID == "" {
		return nil, fmt.Errorf("DISCORD_CLIENT_ID environment variable is required for bot verification")
	}

	channelID := strings.TrimSpace(os.Getenv("DISCORD_CHANNEL_ID"))

	// Validate token format (basic check)
	if len(token) < 50 || !strings.Contains(token, ".") {
		return nil, fmt.Errorf("invalid Discord bot token format")
	}

	// Scraper configuration
	chromePath := os.Getenv("CHROME_PATH")
	if chromePath == "" {
		chromePath = findChromePath()
	}

	userAgent := getEnvOrDefault("USER_AGENT", "Mozilla/5.0 (compatible; FreeGamesBotScraper/2.0; +https://github.com/yourusername/free-games-bot)")

	// Database configuration
	dbPath := getEnvOrDefault("DATABASE_PATH", "games.db")

	// Web configuration
	webPort := getEnvOrDefault("WEB_PORT", ":3000")
	if !strings.HasPrefix(webPort, ":") {
		webPort = ":" + webPort
	}

	// App configuration
	environment := getEnvOrDefault("ENVIRONMENT", "production")
	logLevel := getEnvOrDefault("LOG_LEVEL", "info")

	config := &Config{
		Discord: DiscordConfig{
			Token:           token,
			ClientID:        clientID,
			ChannelID:       channelID,
			MaxRetries:      getEnvInt("DISCORD_MAX_RETRIES", 3),
			RetryDelay:      getEnvDuration("DISCORD_RETRY_DELAY", 5*time.Second),
			CommandTimeout:  getEnvDuration("DISCORD_COMMAND_TIMEOUT", 30*time.Second),
			RateLimitBuffer: getEnvDuration("DISCORD_RATE_LIMIT_BUFFER", 1*time.Second),
		},
		Scraper: ScraperConfig{
			ChromePath:   chromePath,
			UserAgent:    userAgent,
			Timeout:      getEnvDuration("SCRAPER_TIMEOUT", 90*time.Second),
			MaxRetries:   getEnvInt("SCRAPER_MAX_RETRIES", 3),
			RetryDelay:   getEnvDuration("SCRAPER_RETRY_DELAY", 5*time.Second),
			RequestDelay: getEnvDuration("SCRAPER_REQUEST_DELAY", 2*time.Second),
		},
		Database: DatabaseConfig{
			Path:              dbPath,
			MaxConnections:    getEnvInt("DB_MAX_CONNECTIONS", 10),
			ConnectionTimeout: getEnvDuration("DB_CONNECTION_TIMEOUT", 30*time.Second),
			QueryTimeout:      getEnvDuration("DB_QUERY_TIMEOUT", 15*time.Second),
		},
		Web: WebConfig{
			Port:           webPort,
			ReadTimeout:    getEnvDuration("WEB_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getEnvDuration("WEB_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:    getEnvDuration("WEB_IDLE_TIMEOUT", 60*time.Second),
			MaxHeaderBytes: getEnvInt("WEB_MAX_HEADER_BYTES", 1<<20), // 1MB
		},
		App: AppConfig{
			Environment:     environment,
			LogLevel:        logLevel,
			RefreshInterval: getEnvDuration("REFRESH_INTERVAL", 6*time.Hour),
			GracefulTimeout: getEnvDuration("GRACEFUL_TIMEOUT", 30*time.Second),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Discord.Token == "" {
		return fmt.Errorf("discord token is required")
	}

	if c.Discord.ClientID == "" {
		return fmt.Errorf("discord client ID is required")
	}


	if c.Scraper.ChromePath == "" {
		return fmt.Errorf("chrome path not found - please install Chrome/Chromium or set CHROME_PATH")
	}

	if c.App.RefreshInterval < time.Hour {
		return fmt.Errorf("refresh interval must be at least 1 hour to respect Epic Games' servers")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Environment) == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Environment) == "production"
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// findChromePath attempts to find Chrome/Chromium executable
func findChromePath() string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Users\\%USERNAME%\\AppData\\Local\\Google\\Chrome\\Application\\chrome.exe",
		}
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "linux":
		paths = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

