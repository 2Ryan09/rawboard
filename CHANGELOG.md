# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-07-16

### 🎮 Enhanced Arcade Leaderboard System

This major update transforms the leaderboard to show only the highest score per player while preserving all submission history - true to traditional arcade behavior.

### Added

- **Per-Player High Score Tracking**: Leaderboard now shows only the highest score for each three-letter combination
- **Complete Score History**: All score submissions are preserved for analytics and admin access
- **Player Statistics API**: New endpoint to get individual player performance metrics
- **Admin Score History Endpoint**: Protected endpoint to access complete score submission history
- **Automatic Data Migration**: Seamless migration of existing leaderboards to new format
- **Enhanced API Documentation**: Updated examples and use cases for new functionality

### New API Endpoints

- `GET /api/v1/games/{gameId}/players/{initials}/stats` - Get player statistics (public)
- `GET /api/v1/games/{gameId}/scores/all` - Get complete score history (protected)

### New Data Models

- `PlayerStats` - Comprehensive player performance metrics
- `AllScoresRecord` - Complete score submission history
- `PlayerHighScores` - High score tracking per player

### Changed

- **Leaderboard Behavior**: Now displays unique players only (highest score per initials)
- **Score Submission**: All scores are stored; leaderboard shows filtered view
- **Database Schema**: Enhanced storage with backward compatibility
- **Response Structure**: Leaderboard entries now represent unique players

### Enhanced

- **Traditional Arcade Experience**: Matches classic arcade high score table behavior
- **Data Analytics**: Complete submission history enables advanced analytics
- **Player Engagement**: Individual statistics encourage repeat play
- **Admin Capabilities**: Full visibility into game performance and player behavior

### Migration & Compatibility

- ✅ **Automatic Migration**: Existing leaderboards migrate on first access
- ✅ **API Compatibility**: All existing endpoints maintain same interface
- ✅ **Data Preservation**: No existing data is lost during migration
- ✅ **Zero Downtime**: Migration happens transparently during operation

### Technical Details

- **Storage Pattern**: Three-tier storage (all scores, high scores, display leaderboard)
- **Performance**: Optimized for fast leaderboard retrieval with comprehensive history
- **Scalability**: Efficient data structures for large-scale score tracking
- **Memory Usage**: Minimal impact on memory with lazy loading of complete history

### Example Behavior Change

**Before (v1.x)**:

```json
// Leaderboard could show same player multiple times
{
  "entries": [
    { "initials": "AAA", "score": 5000 },
    { "initials": "BBB", "score": 4500 },
    { "initials": "AAA", "score": 4000 }
  ]
}
```

**After (v2.0)**:

```json
// Shows only highest score per player
{
  "entries": [
    { "initials": "AAA", "score": 5000 },
    { "initials": "BBB", "score": 4500 }
  ]
}
// All scores (5000, 4500, 4000) preserved in complete history
```

### Developer Impact

- **Frontend Integration**: Enhanced user experience with per-player tracking
- **Analytics**: Rich data for game performance analysis
- **Backward Compatibility**: No breaking changes to existing integrations
- **New Features**: Optional player statistics and admin endpoints

---

## [1.0.0] - Previous Release

### Initial Release

- Basic leaderboard functionality
- API key authentication
- Redis/Valkey storage
- Health monitoring
- Production deployment support
