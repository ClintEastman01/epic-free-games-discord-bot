# Free Games Bot - Complete Documentation

## 🎮 Overview

The Free Games Bot is a comprehensive Discord bot that automatically monitors Epic Games Store for free games and provides rich notifications to your Discord server. Built with Go, it features a modern architecture with web scraping, database caching, and a complete web-based documentation system.

## 🚀 Quick Start

### 1. Setup the Bot
```bash
# Clone and build
go mod tidy
go build -o free-games-bot cmd/bot/main.go

# Configure environment
cp .env.example .env
# Edit .env with your Discord bot token
```

### 2. Run the Bot
```bash
./free-games-bot
```

### 3. Configure in Discord
```
/setup #your-channel
```

### 4. Access Documentation
Visit: `http://localhost:3000/help`

## 📖 Web Documentation System

The bot now includes a comprehensive web-based documentation system accessible at `/help` endpoint:

### Features:
- **Interactive Navigation**: Tabbed interface with smooth transitions
- **Real-time Statistics**: Live bot status and game counts
- **Complete API Documentation**: All endpoints with examples
- **Responsive Design**: Works on desktop and mobile
- **Copy-to-Clipboard**: Easy command copying
- **Auto-refresh**: Statistics update every 5 minutes

### Available Sections:
1. **Overview** - Bot introduction and key features
2. **Setup Guide** - Step-by-step configuration
3. **Commands** - Complete command reference
4. **Features** - Detailed feature explanations
5. **Architecture** - Technical implementation details
6. **API** - HTTP endpoints documentation
7. **Troubleshooting** - Common issues and solutions

## 🔌 HTTP API Endpoints

### GET /help
Complete interactive documentation interface

### GET /api/status
Returns bot status and statistics:
```json
{
  "status": "online",
  "server_count": 42,
  "game_count": 3,
  "last_update": "2024-01-15T10:30:00Z",
  "uptime": "24/7"
}
```

### GET /api/games
Returns current game statistics:
```json
{
  "free_now": 2,
  "coming_soon": 1,
  "total": 3,
  "last_updated": "2024-01-15T10:30:00Z"
}
```

## 🎯 Discord Commands

### Slash Commands
- `/setup <channel>` - Configure bot (Admin only)
- `/games` - Show current free games
- `/refresh` - Manually refresh games (Admin only)
- `/status` - Show bot status and configuration
- `/help` - Show command help

### Text Commands (in configured channel)
- `!games` or `!freegames` - Show current games
- `!refresh` or `!update` - Refresh games
- `!help` - Show help

## 🏗️ Architecture

### Components:
1. **Discord Bot Layer** (`internal/bot/`) - Discord interactions
2. **Service Layer** (`internal/service/`) - Business logic
3. **Scraper Layer** (`internal/scraper/`) - Web scraping
4. **Database Layer** (`internal/database/`) - Data persistence
5. **Web Layer** (`internal/web/`) - HTTP documentation server

### Data Flow:
1. **Scheduled Trigger** → Every 6 hours
2. **Web Scraping** → Extract game data from Epic Games Store
3. **Data Processing** → Validate and categorize games
4. **Database Storage** → Save to SQLite with deduplication
5. **Change Detection** → Identify new games since last check
6. **Discord Notification** → Send rich embeds to configured channels
7. **Web Documentation** → Serve real-time statistics and docs

## ✨ Key Features

### Automatic Monitoring
- Checks Epic Games Store every 6 hours
- Immediate check on startup
- Retry logic with exponential backoff
- Graceful error handling

### Smart Database System
- SQLite for lightweight persistence
- Duplicate prevention
- Automatic cleanup of old games
- Server configuration storage

### Multi-Server Support
- Per-server channel configuration
- Independent settings per Discord server
- Welcome messages for new servers
- Admin permission checks

### Rich Discord Integration
- Beautiful embed messages with game images
- Color-coded status (Green: Free Now, Blue: Coming Soon)
- Detailed game information
- Slash command support

### Web Documentation
- Complete interactive documentation
- Real-time bot statistics
- API endpoint documentation
- Mobile-responsive design

### Advanced Web Scraping
- Chrome/Chromium browser automation
- JavaScript rendering support
- Cross-platform Chrome detection
- Headless operation for servers

## 🔧 Configuration

### Environment Variables
```env
DISCORD_BOT_TOKEN=your_discord_bot_token_here
DATABASE_PATH=games.db
```

### Bot Permissions Required
- Send Messages
- Use Slash Commands
- Embed Links
- Attach Files

## 🛠️ Development

### Project Structure
```
free-games-scrape/
├── cmd/bot/main.go              # Application entry point
├── internal/
│   ├── app/app.go               # Main application logic
│   ├── bot/discord_bot.go       # Discord bot implementation
│   ├── config/config.go         # Configuration management
│   ├── database/                # Database operations
│   ├── models/                  # Data models
│   ├── scraper/epic_scraper.go  # Web scraping logic
│   ├── service/game_service.go  # Business logic
│   └── web/server.go            # Web documentation server
├── web/
│   ├── static/                  # CSS, JS, images
│   └── templates/               # HTML templates
└── games.db                     # SQLite database
```

### Building and Running
```bash
# Development
go run cmd/bot/main.go

# Production build
go build -o free-games-bot cmd/bot/main.go
./free-games-bot

# With custom port for web server
# (Modify internal/app/app.go to change port)
```

## 🔍 Troubleshooting

### Common Issues

**Bot not responding to commands:**
- Ensure `/setup` has been run
- Check bot permissions
- Verify commands are used in configured channel

**No game notifications:**
- Check if new games are actually available
- Use `/refresh` to manually trigger check
- Verify bot is online and channel exists

**Web documentation not accessible:**
- Check if port 3000 is available
- Ensure firewall allows connections
- Verify web server started successfully

**Scraping failures:**
- Install Chrome/Chromium browser
- Check internet connectivity
- Verify Epic Games Store accessibility

## 📊 Monitoring

### Health Checks
- Web server provides status endpoints
- Database connection monitoring
- Discord bot connection status
- Scraping success/failure tracking

### Logging
- Structured logging with levels
- Error tracking and reporting
- Performance metrics
- User interaction logging

## 🚀 Deployment

### Docker (Recommended)
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o free-games-bot cmd/bot/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates chromium
WORKDIR /root/
COPY --from=builder /app/free-games-bot .
COPY --from=builder /app/web ./web
CMD ["./free-games-bot"]
```

### Systemd Service
```ini
[Unit]
Description=Free Games Discord Bot
After=network.target

[Service]
Type=simple
User=bot
WorkingDirectory=/opt/free-games-bot
ExecStart=/opt/free-games-bot/free-games-bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## 📈 Performance

### Optimizations
- Database indexing for fast queries
- Connection pooling for Discord API
- Efficient memory usage
- Minimal CPU overhead during idle

### Scaling
- Supports unlimited Discord servers
- Horizontal scaling possible
- Database sharding for large deployments
- Load balancing for web documentation

## 🔐 Security

### Best Practices
- Environment variable configuration
- Input validation on all commands
- Rate limiting for user interactions
- Secure token storage
- Minimal required permissions

## 📞 Support

### Getting Help
1. Check the web documentation at `/help`
2. Use `/status` command to check bot health
3. Review logs for error messages
4. Verify configuration and permissions

### Contributing
1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Submit pull request with documentation

---

## 🎉 Conclusion

The Free Games Bot now provides a complete solution for Discord servers wanting automated Epic Games Store notifications, with comprehensive web-based documentation, real-time statistics, and a robust architecture designed for reliability and scalability.

**Key URLs:**
- Documentation: `http://localhost:3000/help`
- Bot Status: `http://localhost:3000/api/status`
- Game Stats: `http://localhost:3000/api/games`

Enjoy never missing a free game again! 🎮