# 🚀 Quick Start Guide

## Prerequisites
- ✅ Go 1.21+ installed
- ✅ .env file configured with LinkedIn credentials
- ✅ Test LinkedIn account (NOT your real account!)

## First Time Setup

```bash
# 1. Navigate to project directory
cd D:\Go\project

# 2. Install dependencies
go mod tidy

# 3. Verify .env file exists and has your credentials
# Check that LINKEDIN_EMAIL and LINKEDIN_PASSWORD are set

# 4. Build the application
go build -o linkedin-bot.exe cmd/bot/main.go
```

## Running the Bot

### Option 1: Run with Go (Recommended for Development)
```bash
go run cmd/bot/main.go
```

### Option 2: Run Compiled Executable
```bash
.\linkedin-bot.exe
```

## What to Expect

### Timeline
- **Initialization**: 10-30 seconds
- **Login**: 30-60 seconds (or instant if session exists)
- **Search**: 1-3 minutes
- **Connection Requests**: 5-15 minutes per request (with cooldowns)

### Total Runtime
- First 10 requests: ~1-2 hours (includes cooldowns)
- Subsequent runs: Faster (session persists)

## Configuration

### Environment Variables (.env)
```env
LINKEDIN_EMAIL=your-email@example.com
LINKEDIN_PASSWORD=your-password

# Search settings
SEARCH_KEYWORDS=Software Engineer
SEARCH_LOCATION=India
MAX_PROFILES_TO_PROCESS=50

# Rate limits
DAILY_CONNECTION_LIMIT=10
CONNECTION_COOLDOWN_MIN=5
CONNECTION_COOLDOWN_MAX=15

# Working hours (24-hour format)
WORK_HOUR_START=9
WORK_HOUR_END=18
```

### What Each Setting Does

- **DAILY_CONNECTION_LIMIT**: Maximum connection requests per day (default: 10)
- **CONNECTION_COOLDOWN_MIN/MAX**: Wait time between requests in minutes (default: 5-15)
- **SEARCH_KEYWORDS**: What profiles to search for
- **SEARCH_LOCATION**: Geographic filter for search
- **MAX_PROFILES_TO_PROCESS**: Maximum profiles to discover (default: 50)
- **WORK_HOUR_START/END**: Active hours (warns if outside, but continues)

## Monitoring

### Watch the Console
The bot provides real-time updates:
- 🔐 Authentication status
- 🔍 Search progress
- 🤝 Connection requests sent
- ⏰ Cooldown timers
- ⚠️  Warnings and errors

### Check the Database
Open `linkedin_bot.db` with any SQLite viewer:
- **profiles**: All discovered profiles
- **connection_requests**: Sent requests
- **messages**: Sent messages (Day 3)
- **sessions**: Saved cookies
- **actions_log**: Complete activity history

### Browser Window
- Browser opens visibly (not headless)
- You can watch the bot interact with LinkedIn
- Human-like mouse movements
- Natural typing speed
- Realistic scrolling

## Troubleshooting

### "Security checkpoint detected"
- **Cause**: LinkedIn requires 2FA or captcha
- **Solution**: 
  1. Complete verification manually in browser
  2. Bot waits 60 seconds automatically
  3. If successful, bot continues
  4. If not, restart bot

### "Login failed - check credentials"
- **Cause**: Wrong email/password or account locked
- **Solution**: 
  1. Verify .env credentials
  2. Try logging in manually first
  3. Check if account is restricted

### "Connect button not found"
- **Cause**: LinkedIn changed DOM structure or profile is unavailable
- **Solution**: Bot logs error and skips to next profile automatically

### "Daily limit reached"
- **Cause**: Sent maximum connection requests for today
- **Solution**: Bot stops automatically. Run again tomorrow.

### Browser crashes
- **Cause**: Network issues or system resources
- **Solution**: 
  1. Restart bot
  2. Session will be restored (no re-login needed)
  3. Already-sent requests won't be duplicated

## Safety Tips

### ⚠️  IMPORTANT WARNINGS

1. **Never use your real LinkedIn account**
   - Create a test account for development
   - Real accounts may be permanently banned

2. **Start with low limits**
   - Begin with DAILY_CONNECTION_LIMIT=5
   - Increase gradually if successful
   - Lower limits = safer

3. **Use realistic cooldowns**
   - Minimum 5 minutes between requests
   - 10-15 minutes is safer
   - LinkedIn detects rapid automation

4. **Don't run continuously**
   - Give LinkedIn time between sessions
   - Run once per day maximum
   - Weekday runs look more realistic

5. **Monitor for checkpoints**
   - If you get 2FA repeatedly, slow down
   - If account is restricted, stop immediately
   - Checkpoints = LinkedIn is suspicious

## Best Practices

### For Demonstration
```bash
# Conservative settings
DAILY_CONNECTION_LIMIT=3
CONNECTION_COOLDOWN_MIN=10
CONNECTION_COOLDOWN_MAX=15
MAX_PROFILES_TO_PROCESS=10
```

### For Demo Video
```bash
# Quick demo (risky but fast)
DAILY_CONNECTION_LIMIT=2
CONNECTION_COOLDOWN_MIN=1
CONNECTION_COOLDOWN_MAX=2
MAX_PROFILES_TO_PROCESS=5
```

### For Testing
1. Run once with DAILY_CONNECTION_LIMIT=1
2. Verify request appears on LinkedIn
3. Check database for logs
4. Increase limit if successful

## Session Management

### First Run
- Bot logs in and saves cookies
- Takes 1-2 minutes for full authentication

### Subsequent Runs
- Bot loads saved cookies
- Skips login (takes 10 seconds)
- Session valid for ~24-48 hours

### When Session Expires
- Bot detects and re-logs in automatically
- No manual intervention needed (unless checkpoint)

## Stopping the Bot

### Graceful Stop
- Let bot finish current connection request
- Bot stops automatically after daily limit
- All progress saved to database

### Emergency Stop
- Press `Ctrl+C` in terminal
- Browser closes automatically
- Session saved (can resume next run)

## Database Queries

### Check sent requests today
```sql
SELECT * FROM connection_requests 
WHERE DATE(sent_at) = DATE('now') 
ORDER BY sent_at DESC;
```

### Count total profiles
```sql
SELECT COUNT(*) FROM profiles;
```

### View recent actions
```sql
SELECT * FROM actions_log 
ORDER BY timestamp DESC 
LIMIT 20;
```

## Next Steps

After successful test run:
1. Review database to verify data
2. Check LinkedIn for sent requests
3. Adjust settings as needed
4. Prepare for Day 3 (Messaging + Documentation)

## Support

For issues:
1. Check DAY2_COMPLETE.txt for known limitations
2. Review logs in actions_log table
3. Verify .env configuration
4. Ensure test account is not restricted

---

**Remember**: This is a proof-of-concept for educational evaluation only. Never use on real LinkedIn accounts or in production environments.
