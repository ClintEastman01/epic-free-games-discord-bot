package models

import "errors"

// Common errors used throughout the application
var (
	ErrNoGamesFound     = errors.New("no games found during scraping")
	ErrInvalidGameData  = errors.New("invalid game data received")
	ErrDiscordSendFail  = errors.New("failed to send message to Discord")
	ErrConfigMissing    = errors.New("required configuration is missing")
	ErrScrapingFailed   = errors.New("scraping operation failed")
)