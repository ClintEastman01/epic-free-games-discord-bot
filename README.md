# Free Games Scraper Discord Bot

A Discord bot that automatically scrapes Epic Games Store for free games and notifies your Discord channel. Now features SQLite database caching and decoupled architecture.

## Features

- Automatic scraping every 6 hours
- SQLite database for game caching
- Interactive Discord commands
- Decoupled bot and scraper
- Smart notifications (only new games)

## Discord Commands

- `!games` or `!freegames` - Show current free games from database
- `!refresh` or `!update` - Manually refresh games from Epic Games Store
- `!help` - Show available commands

## Quick Start

1. **Install dependencies**: `go mod tidy`
2. **Copy environment file**: `cp .env.example .env`
3. **Configure your Discord bot token and channel ID in `.env`**
4. **Run**: `go run cmd/bot/main.go`

## Architecture

The bot is now fully decoupled with these components:

- **Database Layer** (`internal/database/`): SQLite for persistent game caching
- **Service Layer** (`internal/service/`): Business logic for game management  
- **Bot Layer** (`internal/bot/`): Discord interactions with interactive commands
- **Scraper Layer** (`internal/scraper/`): Epic Games Store web scraping

Games are cached in SQLite, so the bot can respond to commands instantly and only sends notifications for truly new games.

## Configuration

Create a `.env` file with:
```env
DISCORD_BOT_TOKEN=your_discord_bot_token_here
YOUR_CHANNEL_ID=your_discord_channel_id_here
DATABASE_PATH=games.db
```

## Building

```bash
# Build the bot
go build -o free-games-bot cmd/bot/main.go

# Run the bot
./free-games-bot
```

The bot will create a SQLite database (`games.db`) automatically and start monitoring for free games!