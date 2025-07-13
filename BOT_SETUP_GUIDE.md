# 🤖 Discord Bot Setup & Invitation Guide

## Step 1: Create Your Discord Bot

### 1.1 Create Discord Application
1. Go to https://discord.com/developers/applications
2. Click "New Application"
3. Name it "Free Games Bot" (or your preferred name)
4. Click "Create"

### 1.2 Create Bot User
1. Go to the "Bot" section in the left sidebar
2. Click "Add Bot"
3. Customize your bot:
   - **Username**: Free Games Bot
   - **Avatar**: Upload a gaming-related image
   - **Public Bot**: ✅ Enable (allows others to invite)
   - **Requires OAuth2 Code Grant**: ❌ Disable

### 1.3 Get Bot Token
1. In the "Bot" section, find the "Token" area
2. Click "Copy" to copy your bot token
3. **IMPORTANT**: Keep this token secret and secure!

## Step 2: Configure Your Bot

### 2.1 Set Up Environment
1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your bot token:
   ```env
   DISCORD_BOT_TOKEN=your_actual_bot_token_here
   DATABASE_PATH=games.db
   ```

### 2.2 Get Your Bot's Client ID
1. In Discord Developer Portal, go to "General Information"
2. Copy the "Application ID" (this is your Client ID)
3. You'll need this for the invite link

## Step 3: Generate Invite Link

### 3.1 Using Discord Developer Portal
1. Go to "OAuth2" → "URL Generator"
2. **Scopes**: Select `bot` and `applications.commands`
3. **Bot Permissions**: Select these:
   - ✅ Send Messages
   - ✅ Use Slash Commands  
   - ✅ Embed Links
   - ✅ Attach Files
   - ✅ Read Message History
   - ✅ Add Reactions

4. Copy the generated URL at the bottom

### 3.2 Manual Invite Link
Replace `YOUR_CLIENT_ID` with your actual Application ID:
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=2147485696&scope=bot%20applications.commands
```

## Step 4: Update Your Bot Code

### 4.1 Add Client ID to Web Server
Edit `internal/web/server.go` and replace `YOUR_BOT_CLIENT_ID` with your actual Client ID:

```go
func (ws *WebServer) handleInvite(w http.ResponseWriter, r *http.Request) {
    // Replace with your actual bot's client ID
    clientID := "123456789012345678"  // Your Application ID here
    permissions := "2147485696"
    // ... rest of the function
}
```

### 4.2 Alternative: Use Environment Variable
Add to your `.env` file:
```env
DISCORD_BOT_TOKEN=your_bot_token_here
DISCORD_CLIENT_ID=your_client_id_here
DATABASE_PATH=games.db
```

Then update the code to read from environment:
```go
import "os"

func (ws *WebServer) handleInvite(w http.ResponseWriter, r *http.Request) {
    clientID := os.Getenv("DISCORD_CLIENT_ID")
    if clientID == "" {
        clientID = "YOUR_BOT_CLIENT_ID" // fallback
    }
    // ... rest of the function
}
```

## Step 5: Start Your Bot

### 5.1 Build and Run
```bash
# Build the bot
go build -o free-games-bot cmd/bot/main.go

# Run the bot
./free-games-bot
```

### 5.2 Verify Bot is Online
- Check console output for "Discord bot is now running"
- Bot should show as online in Discord
- Web server should be available at http://localhost:3000

## Step 6: Invite Bot to Server

### 6.1 Use Your Invite Link
1. Open your generated invite link
2. Select the Discord server
3. Confirm permissions
4. Click "Authorize"

### 6.2 Initial Setup
1. Bot will send a welcome message
2. Run `/setup #channel-name` to configure notifications
3. Test with `/games` to verify it's working

## Step 7: Fix `/setup` Command Issues

If `/setup` isn't working, check these:

### 7.1 Slash Command Registration
The bot automatically registers slash commands on startup. Check logs for:
```
Successfully registered 5 slash commands
```

### 7.2 Permission Issues
- Ensure you have Administrator permissions in Discord
- Bot needs "Use Slash Commands" permission
- Commands may take up to 1 hour to sync globally

### 7.3 Debug Steps
1. Check bot logs for errors
2. Try other commands like `/help` first
3. Re-invite bot if commands don't appear
4. Wait up to 1 hour for global command sync

## Step 8: Troubleshooting

### Common Issues:

**Bot not responding:**
- ✅ Check bot token is correct
- ✅ Ensure bot is online (green status)
- ✅ Verify bot has required permissions

**Slash commands not appearing:**
- ✅ Wait up to 1 hour for global sync
- ✅ Try kicking and re-inviting bot
- ✅ Check "Use Slash Commands" permission

**Setup command fails:**
- ✅ Ensure you have Administrator permissions
- ✅ Check bot can send messages to target channel
- ✅ Verify channel exists and bot can see it

## Step 9: Share Your Bot

### 9.1 Public Bot
If you want others to use your bot:
1. Keep "Public Bot" enabled in Discord Developer Portal
2. Share your invite link publicly
3. Consider hosting on a VPS for 24/7 uptime

### 9.2 Documentation Links
When bot is running, users can access:
- **Documentation**: http://localhost:3000/help
- **Invite Page**: http://localhost:3000/invite
- **Bot Status**: http://localhost:3000/api/status

## Quick Reference

**Required Permissions Value**: `2147485696`
**Invite URL Template**:
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=2147485696&scope=bot%20applications.commands
```

**Essential Commands**:
- `/setup #channel` - Configure bot (Admin only)
- `/games` - Show current free games
- `/status` - Show bot status
- `/help` - Show all commands

That's it! Your Free Games Bot should now be ready to invite and use! 🎮