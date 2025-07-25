# Rawboard API Authentication

## Overview

The Rawboard Arcade API uses API key authentication to protect score submission endpoints while keeping leaderboard viewing public.

## Authentication Method

- **Type**: API Key
- **Required for**: Score submission (POST endpoints)
- **Not required for**: Viewing leaderboards (GET endpoints), health checks

## Setup

### Environment Variable

Set the API key in your environment:

```bash
export RAWBOARD_API_KEY="your-secure-api-key-here"
```

### Development Mode

If no API key is set, authentication is disabled for development:

```bash
# No RAWBOARD_API_KEY set = development mode (no auth required)
go run cmd/server/main.go
```

## API Usage

### Public Endpoints (No Authentication)

```bash
# Get leaderboard - no auth required (shows highest score per player)
curl http://localhost:8080/api/v1/games/tetris/leaderboard

# Get player statistics - no auth required
curl http://localhost:8080/api/v1/games/tetris/players/AAA/stats

# Health check - no auth required
curl http://localhost:8080/health

# Welcome page - no auth required
curl http://localhost:8080/
```

### Protected Endpoints (API Key Required)

#### Submit Score

```bash
# Option 1: X-API-Key Header (Recommended)
curl -X POST http://localhost:8080/api/v1/games/tetris/scores \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"initials": "AAA", "score": 15000}'

# Option 2: Authorization Bearer Header
curl -X POST http://localhost:8080/api/v1/games/tetris/scores \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key-here" \
  -d '{"initials": "BBB", "score": 12000}'
```

#### Get Complete Score History (Admin)

```bash
curl -H "X-API-Key: your-api-key-here" \
     http://localhost:8080/api/v1/games/tetris/scores/all
```

## Next.js Integration

### Environment Variable

```bash
# .env.local
RAWBOARD_API_KEY=your-api-key-here
```

### API Calls

```javascript
// Submit score from Next.js (stores all scores, updates leaderboard with highest)
const submitScore = async (gameId, initials, score) => {
  const response = await fetch(`/api/v1/games/${gameId}/scores`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": process.env.RAWBOARD_API_KEY,
    },
    body: JSON.stringify({ initials, score }),
  });
  return response.json();
};

// Get leaderboard (no auth needed) - shows highest score per player
const getLeaderboard = async (gameId) => {
  const response = await fetch(`/api/v1/games/${gameId}/leaderboard`);
  return response.json();
};

// Get player statistics (no auth needed)
const getPlayerStats = async (gameId, initials) => {
  const response = await fetch(
    `/api/v1/games/${gameId}/players/${initials}/stats`,
  );
  return response.json();
};

// Get complete score history (admin endpoint, requires auth)
const getAllScores = async (gameId) => {
  const response = await fetch(`/api/v1/games/${gameId}/scores/all`, {
    headers: {
      "X-API-Key": process.env.RAWBOARD_API_KEY,
    },
  });
  return response.json();
};
```

## Error Responses

### Missing API Key

```json
{
  "error": "API key required",
  "details": {
    "message": "Please provide API key in X-API-Key header or Authorization: Bearer <key>"
  }
}
```

### Invalid API Key

```json
{
  "error": "Invalid API key"
}
```

## Security Notes

- API key should be kept secret and not exposed in client-side code
- Use environment variables for API key storage
- API key grants access to submit scores for any game
- Consider rotating API keys periodically
- In production, use HTTPS to protect API key in transit

## Migration to JWT

The current API key authentication can easily be upgraded to JWT later:

1. Replace `middleware.APIKeyMiddleware()` with `middleware.JWTMiddleware()`
2. Update client to send JWT tokens instead of API keys
3. All route and handler logic remains the same
