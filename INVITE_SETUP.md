# Discord Bot Invitation Setup Guide

## 🤖 Creating Your Discord Bot

### Step 1: Create a Discord Application
1. Go to https://discord.com/developers/applications
2. Click "New Application"
3. Give your bot a name (e.g., "Free Games Bot")
4. Click "Create"

### Step 2: Create the Bot User
1. In your application, go to the "Bot" section
2. Click "Add Bot"
3. Customize your bot:
   - **Username**: Free Games Bot
   - **Avatar**: Upload a gaming-related image
   - **Public Bot**: ✅ Enabled (so others can invite it)
   - **Requires OAuth2 Code Grant**: ❌ Disabled

### Step 3: Get Your Bot Token
1. In the "Bot" section, under "Token"
2. Click "Copy" to copy your bot token
3. Save this token securely - you'll need it for the `.env` file

### Step 4: Set Bot Permissions
1. Go to the "OAuth2" → "URL Generator" section
2. **Scopes**: Select `bot` and `applications.commands`
3. **Bot Permissions**: Select these permissions:
   - ✅ Send Messages
   - ✅ Use Slash Commands
   - ✅ Embed Links
   - ✅ Attach Files
   - ✅ Read Message History
   - ✅ Add Reactions

### Step 5: Generate Invite Link
The URL Generator will create an invite link like:
```
https://discord.com/api/oauth2/authorize?client_id=YOUR_BOT_ID&permissions=2147485696&scope=bot%20applications.commands
```

## 🚀 Bot Configuration

### Step 6: Configure Environment
1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your bot token:
   ```env
   DISCORD_BOT_TOKEN=your_actual_bot_token_here
   DATABASE_PATH=games.db
   ```

### Step 7: Start Your Bot
```bash
# Build the bot
go build -o free-games-bot cmd/bot/main.go

# Run the bot
./free-games-bot
```

## 📋 Inviting the Bot to Servers

### For Server Owners:
1. Use the invite link generated in Step 5
2. Select the server you want to add the bot to
3. Confirm the permissions
4. Click "Authorize"

### After Invitation:
1. The bot will send a welcome message
2. Run `/setup #channel-name` to configure notifications
3. Test with `/games` to see current free games

## 🔧 Troubleshooting Invitation Issues

### Bot Not Responding:
- ✅ Check bot token is correct in `.env`
- ✅ Ensure bot is online (green status)
- ✅ Verify bot has required permissions
- ✅ Run `/setup` command first

### Permission Errors:
- ✅ Re-invite bot with correct permissions
- ✅ Check channel permissions for the bot
- ✅ Ensure bot role is above other roles (if needed)

### Slash Commands Not Working:
- ✅ Wait up to 1 hour for commands to sync globally
- ✅ Try kicking and re-inviting the bot
- ✅ Check bot has "Use Slash Commands" permission

## 📊 Public Bot Hosting

If you want to make your bot public for others to invite:

### Option 1: Self-Hosted Public Bot
1. Host your bot on a VPS/cloud server
2. Share the invite link publicly
3. Monitor usage and costs

### Option 2: Bot Listing Sites
1. Submit to Discord bot lists like:
   - top.gg
   - discord.bots.gg
   - discordbotlist.com

### Option 3: GitHub Releases
1. Create releases with pre-built binaries
2. Include setup instructions
3. Let users host their own instances

## 🎯 Quick Invite Link Template

Replace `YOUR_BOT_CLIENT_ID` with your actual bot's client ID:

```
https://discord.com/api/oauth2/authorize?client_id=YOUR_BOT_CLIENT_ID&permissions=2147485696&scope=bot%20applications.commands
```

**Required Permissions Value: `2147485696`**
- Send Messages (2048)
- Embed Links (16384)
- Attach Files (32768)
- Read Message History (65536)
- Use Slash Commands (2147483648)
- Add Reactions (64)