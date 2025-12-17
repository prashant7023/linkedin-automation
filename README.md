# LinkedIn Automation PoC

⚠️ **DISCLAIMER: Educational purposes only. This tool violates LinkedIn's Terms of Service. Never use in production or on real accounts.**

## Project Overview

A Go-based LinkedIn automation proof-of-concept demonstrating advanced browser automation, anti-detection techniques, and human-like behavior simulation using the Rod library.

## Features

- ✅ **Authentication System** - Login with session persistence
- ✅ **Profile Search** - Discover and target profiles with pagination
- ✅ **Connection Automation** - Send personalized connection requests
- ✅ **Messaging System** - Follow-up messages for accepted connections
- ✅ **8+ Stealth Techniques** - Advanced anti-detection mechanisms
- ✅ **SQLite Persistence** - Track all actions and state
- ✅ **Activity Scheduling** - Business hours and break simulation

## Anti-Detection Techniques Implemented

### Mandatory (3/3)
1. **Bézier Curve Mouse Movement** - Natural cursor paths with overshoot and micro-corrections
2. **Randomized Timing Patterns** - Variable delays, think time, and action spacing
3. **Browser Fingerprint Masking** - User agent rotation, viewport randomization, webdriver flag removal

### Additional (5/5)
4. **Random Scrolling Behavior** - Variable speed with acceleration/deceleration
5. **Realistic Typing Simulation** - Variable keystroke intervals with occasional typos
6. **Mouse Hovering** - Natural hover delays before clicks
7. **Activity Scheduling** - Business hours operation with break patterns
8. **Rate Limiting** - Connection quotas and cooldown periods

**Total: 8 techniques ✅**

## Project Structure

```
linkedin-automation-poc/
├── cmd/bot/              # Main application entry point
├── internal/
│   ├── browser/          # Browser management & fingerprint masking
│   ├── stealth/          # Human-like behavior engine
│   │   ├── mouse.go      # Bézier curves, overshoot
│   │   ├── typing.go     # Typing simulation
│   │   ├── scroll.go     # Scroll patterns
│   │   └── timing.go     # Delay generators
│   ├── auth/             # Authentication & sessions
│   ├── search/           # Profile discovery
│   ├── connect/          # Connection requests
│   ├── message/          # Messaging automation
│   ├── storage/          # SQLite database layer
│   ├── scheduler/        # Activity timing
│   └── logger/           # Structured logging
├── config/               # Configuration files
└── .env.example          # Environment template
```

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Setup Steps

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd linkedin-automation-poc
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings (DO NOT commit real credentials)
   ```

4. **Review configuration**
   ```bash
   # Edit config/config.yaml to customize behavior
   ```

## Configuration

### Environment Variables (.env)

```env
LINKEDIN_EMAIL=your.email@example.com
LINKEDIN_PASSWORD=your_password
SEARCH_KEYWORDS=Software Engineer
DAILY_CONNECTION_LIMIT=10
DB_PATH=./linkedin_bot.db
```

See [.env.example](.env.example) for all available options.

### Config File (config/config.yaml)

Customize stealth timing, browser settings, message templates, and more. See comments in the file for details.

## Usage

### Run the bot

```bash
go run cmd/bot/main.go
```

### Build executable

```bash
go build -o linkedin-bot.exe cmd/bot/main.go
./linkedin-bot.exe
```

## Database Schema

The bot uses SQLite to persist state across runs:

- **profiles** - Discovered LinkedIn profiles
- **connection_requests** - Sent connection requests
- **messages** - Sent messages
- **sessions** - Browser session cookies
- **actions_log** - Complete activity log

See [internal/storage/store.go](internal/storage/store.go) for full schema.

## Development Status

### ✅ Completed (Day 1)
- Project structure
- Database layer
- Browser setup with fingerprint masking
- Complete stealth engine (mouse, typing, scroll, timing)
- Logger implementation
- Configuration system

### 🚧 In Progress (Day 2)
- Authentication flow
- Profile search
- Connection requests

### 📋 Planned (Day 3)
- Messaging system
- Final polish
- Demo video

## Known Limitations

- **Captcha Detection**: Bot exits gracefully when captcha appears (cannot solve)
- **2FA Required**: Manual intervention needed for two-factor authentication
- **DOM Changes**: LinkedIn frequently updates selectors; may require maintenance
- **Partial Fingerprint Masking**: Some browser fingerprints cannot be fully hidden
- **Detection Possible**: LinkedIn has sophisticated bot detection; this is a PoC only

## Architecture Decisions

### Why SQLite?
- Lightweight, no server required
- ACID transactions for data integrity
- Easy to inspect and debug
- Perfect for single-instance automation

### Why Rod over Selenium?
- Pure Go implementation
- Better performance
- More control over browser internals
- Easier stealth injection

### Why Non-Headless?
- Headless browsers have distinct fingerprints
- More realistic for demonstrating human-like behavior
- Easier to debug and showcase in demo video

## Testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./internal/stealth/...
```

## Contributing

This is an educational project for technical evaluation. Not accepting contributions.

## License

MIT License - See LICENSE file

## Acknowledgments

- [Rod](https://github.com/go-rod/rod) - Go browser automation library
- [SQLite](https://www.sqlite.org/) - Embedded database

## Contact

For questions about this technical assessment, please contact [your-email].

---

**Remember**: This tool is for educational and evaluation purposes only. Using automation tools on LinkedIn violates their Terms of Service and may result in permanent account bans.
