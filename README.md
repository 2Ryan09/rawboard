# 🎮 Rawboard

A traditional arcade-style leaderboard service built with Go and Gin. Perfect for indie games that need fast, reliable leaderboard functionality.

## ✨ Features

- **Enhanced Arcade Leaderboards**: Shows highest score per player (traditional arcade behavior)
- **Complete Score History**: Captures and stores every score submission for analytics
- **Player Statistics**: Individual player performance tracking and statistics
- **REST API**: Simple HTTP endpoints for game integration
- **API Key Authentication**: Secure score submission
- **Redis/Valkey Storage**: Fast, reliable data persistence
- **Health Monitoring**: Built-in health checks and observability
- **Production Ready**: Bugsnag integration, rate limiting, and proper error handling
- **Backward Compatible**: Existing integrations continue to work seamlessly

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Redis/Valkey instance
- (Optional) Docker & Docker Compose

### Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/2Ryan09/rawboard.git
   cd rawboard
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables** (see [Environment Variables](#environment-variables) section)

4. **Start Redis/Valkey** (if using Docker Compose)

   ```bash
   docker-compose up valkey -d
   ```

5. **Run the server**

   ```bash
   go run cmd/server/main.go
   ```

6. **Test the API**
   ```bash
   curl http://localhost:8080/health
   ```

## 🔧 Environment Variables

### Required Variables

| Variable           | Description                                                  | Example                    |
| ------------------ | ------------------------------------------------------------ | -------------------------- |
| `RAWBOARD_API_KEY` | API key for authenticated endpoints (required in production) | `your-secret-api-key-here` |

### Database Configuration

| Variable           | Description                 | Default                  | Example                                                                 |
| ------------------ | --------------------------- | ------------------------ | ----------------------------------------------------------------------- |
| `VALKEY_URI`       | Redis/Valkey connection URI | `redis://localhost:6379` | `redis://user:pass@host:6379/0`<br>`rediss://user:pass@host:port` (SSL) |
| `DATABASE_TIMEOUT` | Database connection timeout | `5s`                     | `10s`, `2m`                                                             |

### Server Configuration

| Variable      | Description         | Default       | Example                 |
| ------------- | ------------------- | ------------- | ----------------------- |
| `PORT`        | Server port         | `8080`        | `3000`, `8000`          |
| `ENVIRONMENT` | Runtime environment | `development` | `production`, `staging` |

### Monitoring & Observability

| Variable          | Description                    | Default      | Example                            |
| ----------------- | ------------------------------ | ------------ | ---------------------------------- |
| `BUGSNAG_API_KEY` | Bugsnag error tracking API key | _(disabled)_ | `94d4ae9e78b0bc3386703e05222adcc3` |

### Leaderboard Configuration

| Variable             | Description                     | Default     | Example      |
| -------------------- | ------------------------------- | ----------- | ------------ |
| `MAX_SCORE_ENTRIES`  | Maximum entries per leaderboard | `10`        | `25`, `100`  |
| `MAX_SCORE_VALUE`    | Maximum allowed score value     | `999999999` | `9999999999` |
| `MAX_GAME_ID_LENGTH` | Maximum game ID string length   | `50`        | `32`, `100`  |

### Testing Variables

| Variable        | Description                   | Purpose                  |
| --------------- | ----------------------------- | ------------------------ |
| `SKIP_DB_TESTS` | Skip database-dependent tests | Set to any value to skip |

## 📁 Example Environment Files

### Development (.env.development)

```bash
# Server
ENVIRONMENT=development
PORT=8080

# Database
VALKEY_URI=redis://localhost:6379
DATABASE_TIMEOUT=5s

# Authentication (optional in development)
RAWBOARD_API_KEY=dev-test-key

# Monitoring (optional)
BUGSNAG_API_KEY=

# Leaderboard Settings
MAX_SCORE_ENTRIES=10
MAX_SCORE_VALUE=999999999
MAX_GAME_ID_LENGTH=50
```

### Production (.env.production)

```bash
# Server
ENVIRONMENT=production
PORT=8080

# Database
VALKEY_URI=redis://your-redis-host:6379
DATABASE_TIMEOUT=10s

# Authentication (REQUIRED)
RAWBOARD_API_KEY=your-secure-api-key-here

# Monitoring
BUGSNAG_API_KEY=your-bugsnag-api-key-here

# Leaderboard Settings
MAX_SCORE_ENTRIES=25
MAX_SCORE_VALUE=9999999999
MAX_GAME_ID_LENGTH=50
```

## 🐳 Docker Setup

### Using Docker Compose (Recommended for Development)

```bash
# Start all services
docker-compose up

# Start only the database
docker-compose up valkey -d

# Build and run the app
docker-compose up app
```

### Using Docker Directly

```bash
# Build the image
docker build -t rawboard .

# Run with environment variables
docker run -d \
  -p 8080:8080 \
  -e ENVIRONMENT=production \
  -e RAWBOARD_API_KEY=your-api-key \
  -e VALKEY_URI=redis://your-redis-host:6379 \
  rawboard
```

## 📚 API Documentation

### Public Endpoints

- `GET /` - API welcome and documentation
- `GET /health` - Health check endpoint
- `GET /api/v1/games/{gameId}/leaderboard` - Get leaderboard (shows highest score per player)
- `GET /api/v1/games/{gameId}/players/{initials}/stats` - Get individual player statistics

### Protected Endpoints (Require API Key)

- `POST /api/v1/games/{gameId}/scores` - Submit new score (stores all scores, updates leaderboard)
- `GET /api/v1/games/{gameId}/scores/all` - Get complete score history (admin endpoint)

### New Leaderboard Behavior

**Enhanced Arcade Experience**: The leaderboard now shows only the **highest score per three-letter combination** while still capturing every score submission. This mirrors traditional arcade behavior where each player appears only once on the high score table.

**Key Changes**:

- 🎯 **Multiple Submissions**: Players can submit multiple scores - all are stored
- 🏆 **Best Score Display**: Leaderboard shows only each player's highest score
- 📊 **Complete History**: All submissions are preserved for analytics
- 📈 **Player Stats**: Individual performance tracking available
- 🔄 **Backward Compatible**: Existing API integrations continue to work

### Example: Multiple Score Behavior

```bash
# Player "AAA" submits multiple scores
curl -H "X-API-Key: your-key" -d '{"initials":"AAA","score":1000}' .../scores
curl -H "X-API-Key: your-key" -d '{"initials":"AAA","score":1500}' .../scores
curl -H "X-API-Key: your-key" -d '{"initials":"AAA","score":1200}' .../scores

# Leaderboard will show: AAA - 1500 (highest score only)
# All scores (1000, 1500, 1200) are stored in complete history
```

#### Authentication

Include your API key in the request header using one of these methods:

**Method 1: X-API-Key Header (Recommended)**

```bash
curl -H "X-API-Key: your-api-key-here" \
     -X POST http://localhost:8080/api/v1/games/my-game/scores
```

**Method 2: Authorization Bearer Header**

```bash
curl -H "Authorization: Bearer your-api-key-here" \
     -X POST http://localhost:8080/api/v1/games/my-game/scores
```

### API Key Setup

#### Development

```bash
# Simple key for testing
export RAWBOARD_API_KEY="dev-test-key"
```

#### Production

```bash
# Generate a secure 32+ character key
export RAWBOARD_API_KEY="$(openssl rand -hex 32)"

# Or create your own secure key
export RAWBOARD_API_KEY="your-super-secure-api-key-minimum-32-chars"
```

**Security Notes:**

- API keys should be at least 32 characters long
- Use cryptographically secure random generation for production
- Store API keys securely (environment variables, secrets management)
- Never commit API keys to version control
- Rotate API keys regularly in production

## 🎮 API Usage Examples

### Submit Score

```bash
curl -X POST http://localhost:8080/api/v1/games/pacman/scores \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"initials": "AAA", "score": 15000}'
```

### Get Leaderboard (Top 10 highest scores per player)

```bash
curl http://localhost:8080/api/v1/games/pacman/leaderboard
```

Response:

```json
{
  "game_id": "pacman",
  "entries": [
    { "initials": "PRO", "score": 50000, "timestamp": "2025-07-16T10:30:00Z" },
    { "initials": "ACE", "score": 45000, "timestamp": "2025-07-16T11:15:00Z" },
    { "initials": "TOP", "score": 40000, "timestamp": "2025-07-16T09:45:00Z" }
  ]
}
```

### Get Player Statistics

```bash
curl http://localhost:8080/api/v1/games/pacman/players/AAA/stats
```

Response:

```json
{
  "initials": "AAA",
  "high_score": 15000,
  "total_scores": 8,
  "last_played": "2025-07-16T14:30:00Z",
  "average_score": 12500.5,
  "first_played": "2025-07-15T16:20:00Z"
}
```

### Get Complete Score History (Admin)

```bash
curl -H "X-API-Key: your-api-key-here" \
     http://localhost:8080/api/v1/games/pacman/scores/all
```

Response:

```json
{
  "game_id": "pacman",
  "scores": [
    { "initials": "AAA", "score": 15000, "timestamp": "2025-07-16T14:30:00Z" },
    { "initials": "AAA", "score": 12000, "timestamp": "2025-07-16T13:15:00Z" },
    { "initials": "BBB", "score": 18000, "timestamp": "2025-07-16T12:45:00Z" },
    { "initials": "AAA", "score": 10000, "timestamp": "2025-07-16T11:30:00Z" }
  ],
  "updated": "2025-07-16T14:30:00Z"
}
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Skip database tests (if Redis unavailable)
SKIP_DB_TESTS=1 go test ./...
```

## 🔄 Migration & Backward Compatibility

### Automatic Migration

The enhanced leaderboard system automatically migrates existing data:

1. **First Request**: When an existing game's leaderboard is accessed, it's automatically migrated
2. **Data Preservation**: All existing scores are preserved in the new format
3. **No Downtime**: Migration happens transparently during normal operation
4. **API Compatibility**: All existing API calls continue to work exactly as before

### Enhanced Behavior Details

**Before Enhancement**:

- Leaderboard could show same player multiple times
- Only top 10 scores stored (older scores lost)
- No player-specific statistics available

**After Enhancement**:

- Leaderboard shows each player's highest score only (max 10 unique players)
- All scores preserved in complete history
- Individual player statistics available
- Admin endpoint for accessing complete score data

### Migration Process

```bash
# Existing leaderboard structure:
{
  "game_id": "pacman",
  "entries": [
    {"initials": "AAA", "score": 5000},
    {"initials": "BBB", "score": 4000},
    {"initials": "AAA", "score": 3000}  // duplicate player
  ]
}

# After automatic migration:
# 1. Complete history stored in all_scores:pacman
# 2. Player high scores tracked in player_high_scores:pacman
# 3. Leaderboard filtered to show: AAA(5000), BBB(4000)
```

## 🔧 Development

### Project Structure

```
rawboard/
├── cmd/                    # Application entrypoints
│   ├── server/            # Main server application
│   └── test-db/           # Database testing utility
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database interface and implementations
│   ├── handlers/          # HTTP request handlers
│   ├── leaderboard/       # Leaderboard business logic
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models
│   └── repository/        # Data access layer
├── api/                   # API documentation
├── migrations/            # Database migrations
└── pkg/                   # Public packages
```

### Building

```bash
# Build for current platform
go build -o rawboard cmd/server/main.go

# Build for Linux (production)
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rawboard cmd/server/main.go
```

## 🚀 Deployment

### Production Checklist

- [ ] Set `ENVIRONMENT=production`
- [ ] Configure secure `RAWBOARD_API_KEY`
- [ ] Set up Redis/Valkey with persistence
- [ ] Configure `BUGSNAG_API_KEY` for error tracking
- [ ] Adjust leaderboard limits if needed
- [ ] Set up proper database backups
- [ ] Configure reverse proxy (nginx, etc.)
- [ ] Set up SSL/TLS certificates

### Environment-Specific Behavior

**Development Mode** (`ENVIRONMENT=development`):

- API key authentication is optional
- Continues running even if database is unavailable
- Detailed error messages
- Debug logging enabled

**Production Mode** (`ENVIRONMENT=production`):

- API key authentication is required
- Exits if database connection fails
- Minimal error exposure
- Optimized logging

## 📄 License

[Add your license here]

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 🐛 Troubleshooting

### Common Issues

**Database Connection Failed**

- Ensure Redis/Valkey is running
- Check `VALKEY_URI` format
- Verify network connectivity
- Check firewall settings

**API Key Authentication Issues**

- Ensure `RAWBOARD_API_KEY` is set in production
- Include `X-API-Key` header in requests
- Verify the API key matches exactly

**Health Check Failing**

- Check if server is running on correct port
- Verify no port conflicts
- Check application logs for errors

**Leaderboard Migration Issues**

- Migration happens automatically on first access
- Check logs for migration-related messages
- Verify database has sufficient storage space
- Ensure Redis/Valkey has persistence enabled

**Player Statistics Not Available**

- Player must have submitted at least one score
- Check that the game ID and initials are correct
- Verify the game has been accessed since the enhancement (triggers migration)

**Score History Missing**

- Complete history only available for scores submitted after migration
- Pre-migration scores are preserved but marked as legacy data
- Use the admin endpoint `/scores/all` to access complete history

### Getting Help

- Check the [Issues](https://github.com/2Ryan09/rawboard/issues) page
- Review application logs for migration messages
- Verify environment variable configuration
- Test with the new `/players/{initials}/stats` endpoint for debugging

## 📋 Quick Reference

### New Enhanced Behavior Summary

- 🎯 **Multiple Scores**: Submit as many scores as you want per player
- 🏆 **Best Display**: Leaderboard shows only highest score per initials
- 📊 **Complete History**: Every score stored for analytics
- 📈 **Player Stats**: Individual performance tracking
- 🔄 **Auto Migration**: Existing data upgraded automatically
- 📡 **Same API**: All existing code continues to work

### Endpoint Quick Reference

```bash
# Submit score (stores all, shows best)
POST /api/v1/games/{gameId}/scores                    [Protected]

# Get leaderboard (unique players only)
GET  /api/v1/games/{gameId}/leaderboard               [Public]

# Get player statistics
GET  /api/v1/games/{gameId}/players/{initials}/stats  [Public]

# Get complete score history
GET  /api/v1/games/{gameId}/scores/all                [Protected]
```

### Key Files Changed

- `internal/leaderboard/service.go` - Enhanced score submission logic
- `internal/models/player.go` - New data models for tracking
- `internal/handlers/leaderboard.go` - New API endpoints
- `internal/handlers/routes.go` - Updated routing configuration
