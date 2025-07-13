# Epic Games Free Games Discord Bot

A Discord bot that automatically monitors Epic Games Store for free games and posts updates to your Discord server. Never miss a free game again!

## Features

- **Automatic Monitoring**: Checks Epic Games Store every 24 hours for free games
- **Real-time Notifications**: Posts updates about currently free games and upcoming free games
- **Rich Discord Embeds**: Beautiful formatted messages with game images and details
- **Cross-Platform Support**: Automatically detects Chrome/Chromium on Windows, macOS, and Linux
- **Modular Architecture**: Clean, maintainable code structure following Go best practices
- **Smart Scheduling**: Continues monitoring even when no active free games remain
- **Detailed Game Info**: Includes game titles, images, and availability periods
- **Reliable Scraping**: Uses Chrome automation with retry logic for consistent data collection
- **Error Handling**: Comprehensive error handling with Discord notifications
- **Easy Configuration**: Environment-based configuration with sensible defaults

## Quick Start

### Prerequisites

- Go 1.24.5 or later
- Chrome/Chromium browser installed
- Discord Bot Token
- Discord Server with appropriate permissions

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd free-games-scrape
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   
   Create a `.env` file in the project root:
   ```env
   DISCORD_BOT_TOKEN=your_bot_token_here
   YOUR_CHANNEL_ID=your_discord_channel_id_here
   ```

4. **Configure Chrome path** (if needed)
   
   The bot is currently configured for macOS Chrome. For other systems, update the Chrome path in `main.go`:
   ```go
   // For Linux
   chromedp.ExecPath("/usr/bin/google-chrome"),
   
   // For Windows
   chromedp.ExecPath("C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"),
   ```

5. **Run the bot**
   ```bash
   # Using the new modular structure (recommended)
   go run cmd/bot/main.go
   
   # Or using the Makefile
   make run
   
   # Or build and run the binary
   make build
   ./bin/epic-games-bot
   ```

## Configuration

### Discord Bot Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application and bot
3. Copy the bot token to your `.env` file
4. Invite the bot to your server with these permissions:
   - Send Messages
   - Read Message History
   - View Channels

### Getting Channel ID

1. Enable Developer Mode in Discord (User Settings → Advanced → Developer Mode)
2. Right-click on your desired channel
3. Select "Copy ID"
4. Paste the ID in your `.env` file

## How It Works

1. **Initial Check**: Bot performs an immediate check when started
2. **Periodic Monitoring**: Checks every 24 hours for updates
3. **Game Classification**: Separates games into "Free Now" and "Coming Soon" categories
4. **Discord Updates**: Posts formatted messages with game details
5. **Smart Shutdown**: Automatically stops when no active free games remain

## Game Information Provided

For each free game, the bot provides:
- **Game Title**: Official name of the game
- **Game Image**: Promotional artwork/thumbnail
- **Availability Status**: "Free Now" or "Coming Soon"
- **Free Period**: When the game is/will be available for free

## Project Structure

```
free-games-scrape/
├── cmd/
│   └── bot/
│       └── main.go          # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go           # Main application logic
│   ├── bot/
│   │   └── discord_bot.go   # Discord bot implementation
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── models/
│   │   └── game.go          # Game data models
│   └── scraper/
│       └── epic_scraper.go  # Epic Games Store scraper
├── go.mod                   # Go module dependencies
├── go.sum                   # Dependency checksums
├── Makefile                 # Build and development commands
├── .env                     # Environment variables (create this)
├── .gitignore              # Git ignore rules
├── TODO.md                 # Development roadmap
├── README.md               # This file
└── main.go                 # Deprecated (use cmd/bot/main.go)
```

## Dependencies

- **[discordgo](https://github.com/bwmarrin/discordgo)**: Discord API wrapper for Go
- **[chromedp](https://github.com/chromedp/chromedp)**: Chrome DevTools Protocol for web scraping
- **[godotenv](https://github.com/joho/godotenv)**: Environment variable management

## Future Features

This bot is actively being developed! Check out [TODO.md](TODO.md) for planned features including:

- **Multi-server support** with per-server configuration
- **Slash commands** for better Discord integration
- **Rich embeds** with enhanced formatting and images
- **Notification preferences** and customization options
- **Game tracking** to avoid duplicate notifications
- **Web dashboard** for easy configuration

## Troubleshooting

### Common Issues

**Bot doesn't start:**
- Verify your Discord bot token is correct
- Ensure the channel ID is valid
- Check that Chrome/Chromium is installed and accessible

**No messages posted:**
- Confirm bot has permission to send messages in the target channel
- Check bot logs for error messages
- Verify Epic Games Store is accessible from your network

**Scraping fails:**
- Update Chrome path in the code if using a different OS
- Check if Epic Games Store layout has changed
- Ensure stable internet connection

### Logs

The bot provides detailed logging to help diagnose issues:
- Startup and shutdown events
- Scraping attempts and results
- Discord API interactions
- Error messages with context

## Contributing

Contributions are welcome! Please feel free to:
- Report bugs and issues
- Suggest new features
- Submit pull requests
- Improve documentation

## License

This project is open source. Please ensure you comply with Epic Games Store's terms of service when using this bot.

## Disclaimer

This bot is for educational and personal use only. It scrapes publicly available information from Epic Games Store. Please respect Epic Games' terms of service and rate limits. The developers are not responsible for any misuse of this software.

---

**Made with love for the gaming community**

*Last updated: January 2025*