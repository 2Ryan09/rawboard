package leaderboard

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"rawboard/internal/database"
	"rawboard/internal/models"
)

// Service handles leaderboard operations
type Service struct {
	db database.DB
}

// NewService creates a new leaderboard service
func NewService(db database.DB) *Service {
	return &Service{db: db}
}

// SubmitScore submits a new score entry (traditional arcade style)
// Now stores all scores and maintains per-player high scores
func (s *Service) SubmitScore(ctx context.Context, gameID, initials string, score int64) error {
	// Validate initials (should be 3 characters, no spaces allowed)
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 || strings.Contains(initials, " ") {
		return fmt.Errorf("initials must be exactly 3 characters with no spaces")
	}

	// Store the score in all scores history
	if err := s.addToAllScores(ctx, gameID, initials, score); err != nil {
		return fmt.Errorf("failed to store score in history: %w", err)
	}

	// Update player's high score if necessary
	if err := s.updatePlayerHighScore(ctx, gameID, initials, score); err != nil {
		return fmt.Errorf("failed to update player high score: %w", err)
	}

	// Regenerate the filtered leaderboard
	return s.regenerateFilteredLeaderboard(ctx, gameID)
}

// submitScoreAtomic uses Redis sorted sets for efficient score management
func (s *Service) submitScoreAtomic(ctx context.Context, gameID, initials string, score int64) error {
	// Create unique member key with timestamp to handle duplicate scores
	timestamp := time.Now().UnixNano()
	member := fmt.Sprintf("%s:%d", initials, timestamp)
	key := fmt.Sprintf("leaderboard:%s", gameID)

	// Add to sorted set (Redis will maintain order automatically)
	if err := s.db.Set(ctx, fmt.Sprintf("%s:member:%s", key, member), fmt.Sprintf(`{"initials":"%s","score":%d,"timestamp":%d}`, initials, score, timestamp)); err != nil {
		return fmt.Errorf("failed to store score entry: %w", err)
	}

	// For now, fall back to the original method to maintain compatibility
	// TODO: Implement full Redis sorted set operations in the database interface
	return s.submitScoreLegacy(ctx, gameID, initials, score)
}

// submitScoreLegacy maintains the original implementation for compatibility
func (s *Service) submitScoreLegacy(ctx context.Context, gameID, initials string, score int64) error {
	// Create the score entry
	entry := models.ScoreEntry{
		Initials:  initials,
		Score:     score,
		Timestamp: time.Now(),
	}

	// Get current leaderboard
	leaderboard, err := s.GetLeaderboard(ctx, gameID)
	if err != nil {
		// If no leaderboard exists yet, create a new one
		leaderboard = &models.Leaderboard{
			GameID:  gameID,
			Entries: []models.ScoreEntry{},
		}
	}

	// Pre-allocate slice with capacity to avoid reallocations
	if cap(leaderboard.Entries) < len(leaderboard.Entries)+1 {
		newEntries := make([]models.ScoreEntry, len(leaderboard.Entries), len(leaderboard.Entries)+10)
		copy(newEntries, leaderboard.Entries)
		leaderboard.Entries = newEntries
	}

	// Add new entry
	leaderboard.Entries = append(leaderboard.Entries, entry)

	// Sort by score (highest first) - use stable sort for consistent ordering
	sort.SliceStable(leaderboard.Entries, func(i, j int) bool {
		if leaderboard.Entries[i].Score == leaderboard.Entries[j].Score {
			// If scores are equal, newer entries come first (traditional arcade behavior)
			return leaderboard.Entries[i].Timestamp.After(leaderboard.Entries[j].Timestamp)
		}
		return leaderboard.Entries[i].Score > leaderboard.Entries[j].Score
	})

	// Keep only top 10 scores (traditional arcade limit)
	if len(leaderboard.Entries) > 10 {
		leaderboard.Entries = leaderboard.Entries[:10]
	}

	// Save back to database
	return s.saveLeaderboard(ctx, leaderboard)
}

// GetLeaderboard returns the current leaderboard for a game
// This now returns the filtered leaderboard (highest score per player)
func (s *Service) GetLeaderboard(ctx context.Context, gameID string) (*models.Leaderboard, error) {
	key := fmt.Sprintf("leaderboard:%s", gameID)

	data, err := s.db.Get(ctx, key)
	if err != nil {
		// Try to migrate existing data if this is a legacy leaderboard
		if migrateErr := s.MigrateExistingLeaderboard(ctx, gameID); migrateErr != nil {
			return nil, fmt.Errorf("no leaderboard found for game and migration failed: %w", migrateErr)
		}

		// Try again after migration
		data, err = s.db.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("no leaderboard found for game")
		}
	}

	var leaderboard models.Leaderboard
	// Use a decoder with pre-allocated buffer for better memory efficiency
	decoder := json.NewDecoder(strings.NewReader(data))
	if err := decoder.Decode(&leaderboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal leaderboard: %w", err)
	}

	return &leaderboard, nil
}

// saveLeaderboard saves a leaderboard to the database with optimized encoding
func (s *Service) saveLeaderboard(ctx context.Context, leaderboard *models.Leaderboard) error {
	// Use buffer pool to reduce allocations
	var buf strings.Builder
	buf.Grow(1024) // Pre-allocate reasonable size for typical leaderboard JSON

	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // Slight performance improvement for non-HTML content

	if err := encoder.Encode(leaderboard); err != nil {
		return fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	key := fmt.Sprintf("leaderboard:%s", leaderboard.GameID)
	// Remove trailing newline that encoder.Encode adds
	jsonData := strings.TrimSuffix(buf.String(), "\n")
	return s.db.Set(ctx, key, jsonData)
}

// addToAllScores adds a score entry to the complete score history
func (s *Service) addToAllScores(ctx context.Context, gameID, initials string, score int64) error {
	key := fmt.Sprintf("all_scores:%s", gameID)

	// Create the score entry
	entry := models.ScoreEntry{
		Initials:  initials,
		Score:     score,
		Timestamp: time.Now(),
	}

	// Get existing all scores record
	allScores, err := s.getAllScores(ctx, gameID)
	if err != nil {
		// If no record exists yet, create a new one
		allScores = &models.AllScoresRecord{
			GameID:  gameID,
			Scores:  []models.ScoreEntry{},
			Updated: time.Now(),
		}
	}

	// Add new entry
	allScores.Scores = append(allScores.Scores, entry)
	allScores.Updated = time.Now()

	// Save back to database
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(allScores); err != nil {
		return fmt.Errorf("failed to marshal all scores: %w", err)
	}

	jsonData := strings.TrimSuffix(buf.String(), "\n")
	return s.db.Set(ctx, key, jsonData)
}

// updatePlayerHighScore updates a player's high score if the new score is higher
func (s *Service) updatePlayerHighScore(ctx context.Context, gameID, initials string, score int64) error {
	key := fmt.Sprintf("player_high_scores:%s", gameID)

	// Get existing high scores
	highScores, err := s.getPlayerHighScores(ctx, gameID)
	if err != nil {
		// If no record exists yet, create a new one
		highScores = &models.PlayerHighScores{
			GameID:     gameID,
			HighScores: make(map[string]models.ScoreEntry),
			Updated:    time.Now(),
		}
	}

	// Check if this is a new high score for the player
	existingEntry, exists := highScores.HighScores[initials]
	if !exists || score > existingEntry.Score {
		// Update or create the high score entry
		highScores.HighScores[initials] = models.ScoreEntry{
			Initials:  initials,
			Score:     score,
			Timestamp: time.Now(),
		}
		highScores.Updated = time.Now()

		// Save back to database
		var buf strings.Builder
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(highScores); err != nil {
			return fmt.Errorf("failed to marshal high scores: %w", err)
		}

		jsonData := strings.TrimSuffix(buf.String(), "\n")
		return s.db.Set(ctx, key, jsonData)
	}

	return nil // No update needed
}

// regenerateFilteredLeaderboard creates a leaderboard showing only the highest score per initials
func (s *Service) regenerateFilteredLeaderboard(ctx context.Context, gameID string) error {
	// Get all player high scores
	highScores, err := s.getPlayerHighScores(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get player high scores: %w", err)
	}

	// Convert map to slice for sorting
	entries := make([]models.ScoreEntry, 0, len(highScores.HighScores))
	for _, entry := range highScores.HighScores {
		entries = append(entries, entry)
	}

	// Sort by score (highest first) - use stable sort for consistent ordering
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			// If scores are equal, newer entries come first (traditional arcade behavior)
			return entries[i].Timestamp.After(entries[j].Timestamp)
		}
		return entries[i].Score > entries[j].Score
	})

	// Keep only top 10 scores
	if len(entries) > 10 {
		entries = entries[:10]
	}

	// Create the filtered leaderboard
	leaderboard := &models.Leaderboard{
		GameID:  gameID,
		Entries: entries,
	}

	// Save the filtered leaderboard
	return s.saveLeaderboard(ctx, leaderboard)
}

// getAllScores retrieves the complete score history for a game
func (s *Service) getAllScores(ctx context.Context, gameID string) (*models.AllScoresRecord, error) {
	key := fmt.Sprintf("all_scores:%s", gameID)

	data, err := s.db.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("no score history found for game")
	}

	var allScores models.AllScoresRecord
	decoder := json.NewDecoder(strings.NewReader(data))
	if err := decoder.Decode(&allScores); err != nil {
		return nil, fmt.Errorf("failed to unmarshal all scores: %w", err)
	}

	return &allScores, nil
}

// getPlayerHighScores retrieves the high scores for all players in a game
func (s *Service) getPlayerHighScores(ctx context.Context, gameID string) (*models.PlayerHighScores, error) {
	key := fmt.Sprintf("player_high_scores:%s", gameID)

	data, err := s.db.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("no player high scores found for game")
	}

	var highScores models.PlayerHighScores
	decoder := json.NewDecoder(strings.NewReader(data))
	if err := decoder.Decode(&highScores); err != nil {
		return nil, fmt.Errorf("failed to unmarshal player high scores: %w", err)
	}

	return &highScores, nil
}

// getRawLeaderboard gets the raw leaderboard data without triggering migration logic
// This is used internally to avoid infinite recursion during migration
func (s *Service) getRawLeaderboard(ctx context.Context, gameID string) (*models.Leaderboard, error) {
	key := fmt.Sprintf("leaderboard:%s", gameID)

	data, err := s.db.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("no raw leaderboard found for game: %w", err)
	}

	var leaderboard models.Leaderboard
	// Use a decoder with pre-allocated buffer for better memory efficiency
	decoder := json.NewDecoder(strings.NewReader(data))
	if err := decoder.Decode(&leaderboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw leaderboard: %w", err)
	}

	return &leaderboard, nil
}

// GetPlayerStats returns comprehensive statistics for a specific player
func (s *Service) GetPlayerStats(ctx context.Context, gameID, initials string) (*models.PlayerStats, error) {
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 {
		return nil, fmt.Errorf("initials must be exactly 3 characters")
	}

	// Get all scores to calculate statistics
	allScores, err := s.getAllScores(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get score history: %w", err)
	}

	// Filter scores for this player
	playerScores := make([]models.ScoreEntry, 0)
	for _, entry := range allScores.Scores {
		if entry.Initials == initials {
			playerScores = append(playerScores, entry)
		}
	}

	if len(playerScores) == 0 {
		return nil, fmt.Errorf("no scores found for player %s", initials)
	}

	// Calculate statistics
	var highScore int64
	var totalScore int64
	var firstPlayed, lastPlayed time.Time

	for i, entry := range playerScores {
		if entry.Score > highScore {
			highScore = entry.Score
		}
		totalScore += entry.Score

		if i == 0 {
			firstPlayed = entry.Timestamp
			lastPlayed = entry.Timestamp
		} else {
			if entry.Timestamp.Before(firstPlayed) {
				firstPlayed = entry.Timestamp
			}
			if entry.Timestamp.After(lastPlayed) {
				lastPlayed = entry.Timestamp
			}
		}
	}

	averageScore := float64(totalScore) / float64(len(playerScores))

	return &models.PlayerStats{
		Initials:     initials,
		HighScore:    highScore,
		TotalScores:  len(playerScores),
		LastPlayed:   lastPlayed,
		AverageScore: averageScore,
		FirstPlayed:  firstPlayed,
	}, nil
}

// GetAllScoresForGame returns all scores submitted for a game (for admin/analytics)
func (s *Service) GetAllScoresForGame(ctx context.Context, gameID string) (*models.AllScoresRecord, error) {
	return s.getAllScores(ctx, gameID)
}

// calculateAchievements determines which achievements a player has unlocked
func (s *Service) calculateAchievements(playerScores []models.ScoreEntry, highScore int64) []models.Achievement {
	achievements := make([]models.Achievement, 0)

	if len(playerScores) == 0 {
		return achievements
	}

	// Sort scores by timestamp for achievement calculation
	sort.Slice(playerScores, func(i, j int) bool {
		return playerScores[i].Timestamp.Before(playerScores[j].Timestamp)
	})

	firstScore := playerScores[0]

	// First Score Achievement
	achievements = append(achievements, models.Achievement{
		ID:          "first_score",
		Name:        "First Score",
		Description: "Submit your first score",
		UnlockedAt:  firstScore.Timestamp,
		Icon:        "🎯",
	})

	// Score milestone achievements
	milestones := []struct {
		score int64
		id    string
		name  string
		icon  string
	}{
		{1000, "score_1k", "Getting Started", "⭐"},
		{5000, "score_5k", "Rising Star", "🌟"},
		{10000, "score_10k", "High Achiever", "💫"},
		{25000, "score_25k", "Score Master", "🏆"},
		{50000, "score_50k", "Legend", "👑"},
	}

	for _, milestone := range milestones {
		if highScore >= milestone.score {
			// Find when this milestone was first achieved
			var unlockedAt time.Time
			for _, score := range playerScores {
				if score.Score >= milestone.score {
					unlockedAt = score.Timestamp
					break
				}
			}

			achievements = append(achievements, models.Achievement{
				ID:          milestone.id,
				Name:        milestone.name,
				Description: fmt.Sprintf("Reach %d points", milestone.score),
				UnlockedAt:  unlockedAt,
				Icon:        milestone.icon,
			})
		}
	}

	// Dedication achievements
	if len(playerScores) >= 5 {
		achievements = append(achievements, models.Achievement{
			ID:          "dedicated_player",
			Name:        "Dedicated Player",
			Description: "Submit 5 or more scores",
			UnlockedAt:  playerScores[4].Timestamp, // 5th score
			Icon:        "🎮",
		})
	}

	if len(playerScores) >= 10 {
		achievements = append(achievements, models.Achievement{
			ID:          "score_hunter",
			Name:        "Score Hunter",
			Description: "Submit 10 or more scores",
			UnlockedAt:  playerScores[9].Timestamp, // 10th score
			Icon:        "🏹",
		})
	}

	return achievements
}

// GetEnhancedPlayerStats returns comprehensive statistics with achievements
func (s *Service) GetEnhancedPlayerStats(ctx context.Context, gameID, initials string, includeHistory bool) (*models.EnhancedPlayerStats, error) {
	initials = strings.ToUpper(strings.TrimSpace(initials))
	if len(initials) != 3 {
		return nil, fmt.Errorf("initials must be exactly 3 characters")
	}

	// Get all scores to calculate statistics
	allScores, err := s.getAllScores(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get score history: %w", err)
	}

	// Filter scores for this player
	playerScores := make([]models.ScoreEntry, 0)
	for _, entry := range allScores.Scores {
		if entry.Initials == initials {
			playerScores = append(playerScores, entry)
		}
	}

	if len(playerScores) == 0 {
		return nil, fmt.Errorf("no scores found for player %s", initials)
	}

	// Calculate basic statistics
	var highScore int64
	var totalScore int64
	var firstPlayed, lastPlayed time.Time

	for i, entry := range playerScores {
		if entry.Score > highScore {
			highScore = entry.Score
		}
		totalScore += entry.Score

		if i == 0 {
			firstPlayed = entry.Timestamp
			lastPlayed = entry.Timestamp
		} else {
			if entry.Timestamp.Before(firstPlayed) {
				firstPlayed = entry.Timestamp
			}
			if entry.Timestamp.After(lastPlayed) {
				lastPlayed = entry.Timestamp
			}
		}
	}

	averageScore := float64(totalScore) / float64(len(playerScores))

	// Get current rank from leaderboard
	var currentRank *int
	leaderboard, err := s.GetLeaderboard(ctx, gameID)
	if err == nil {
		for i, entry := range leaderboard.Entries {
			if entry.Initials == initials {
				rank := i + 1
				currentRank = &rank
				break
			}
		}
	}

	// Calculate achievements
	achievements := s.calculateAchievements(playerScores, highScore)

	// Prepare score history if requested
	var scoreHistory []models.ScoreEntry
	if includeHistory {
		scoreHistory = playerScores
	}

	return &models.EnhancedPlayerStats{
		Initials:     initials,
		HighScore:    highScore,
		TotalScores:  len(playerScores),
		LastPlayed:   lastPlayed,
		AverageScore: averageScore,
		FirstPlayed:  firstPlayed,
		CurrentRank:  currentRank,
		Achievements: achievements,
		ScoreHistory: scoreHistory,
	}, nil
}

// GetScoreAnalysis returns comprehensive analysis for a game
func (s *Service) GetScoreAnalysis(ctx context.Context, gameID string, topPlayersLimit int) (*models.ScoreAnalysisResponse, error) {
	// Get all scores
	allScores, err := s.getAllScores(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get score history: %w", err)
	}

	if len(allScores.Scores) == 0 {
		return nil, fmt.Errorf("no scores found for game")
	}

	// Calculate basic statistics
	totalScores := len(allScores.Scores)
	var highestScore int64
	var totalScore int64
	var lastActivity time.Time
	playerMap := make(map[string][]models.ScoreEntry)

	// Group scores by player and calculate totals
	for _, score := range allScores.Scores {
		if score.Score > highestScore {
			highestScore = score.Score
		}
		totalScore += score.Score

		if score.Timestamp.After(lastActivity) {
			lastActivity = score.Timestamp
		}

		playerMap[score.Initials] = append(playerMap[score.Initials], score)
	}

	totalPlayers := len(playerMap)
	averageScore := float64(totalScore) / float64(totalScores)

	// Get top players with enhanced stats
	topPlayers := make([]models.EnhancedPlayerStats, 0)
	leaderboard, _ := s.GetLeaderboard(ctx, gameID)

	limit := topPlayersLimit
	if limit <= 0 || limit > 10 {
		limit = 10
	}

	for i, entry := range leaderboard.Entries {
		if i >= limit {
			break
		}

		enhancedStats, err := s.GetEnhancedPlayerStats(ctx, gameID, entry.Initials, false)
		if err == nil {
			topPlayers = append(topPlayers, *enhancedStats)
		}
	}

	// Calculate score distribution
	scoreDistribution := make(map[string]int)
	ranges := []struct {
		min, max int64
		label    string
	}{
		{0, 999, "0-999"},
		{1000, 4999, "1K-5K"},
		{5000, 9999, "5K-10K"},
		{10000, 24999, "10K-25K"},
		{25000, 49999, "25K-50K"},
		{50000, 999999999, "50K+"},
	}

	for _, score := range allScores.Scores {
		for _, r := range ranges {
			if score.Score >= r.min && score.Score <= r.max {
				scoreDistribution[r.label]++
				break
			}
		}
	}

	// Get recent achievements (last 24 hours)
	recentAchievements := make([]models.Achievement, 0)
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, playerScores := range playerMap {
		// Get player's highest score
		var highScore int64
		for _, score := range playerScores {
			if score.Score > highScore {
				highScore = score.Score
			}
		}

		achievements := s.calculateAchievements(playerScores, highScore)
		for _, achievement := range achievements {
			if achievement.UnlockedAt.After(cutoff) {
				recentAchievements = append(recentAchievements, achievement)
			}
		}
	}

	return &models.ScoreAnalysisResponse{
		GameID:             gameID,
		TotalPlayers:       totalPlayers,
		TotalScores:        totalScores,
		HighestScore:       highestScore,
		AverageScore:       averageScore,
		LastActivity:       lastActivity,
		TopPlayers:         topPlayers,
		ScoreDistribution:  scoreDistribution,
		RecentAchievements: recentAchievements,
		Updated:            time.Now(),
	}, nil
}

// MigrateExistingLeaderboard migrates an existing leaderboard to the new storage format
// This should be called for games that have existing leaderboards before the new system
func (s *Service) MigrateExistingLeaderboard(ctx context.Context, gameID string) error {
	// Get existing leaderboard data directly without triggering migration recursion
	leaderboard, err := s.getRawLeaderboard(ctx, gameID)
	if err != nil {
		// If no leaderboard exists, nothing to migrate
		return nil
	}

	// Check if already migrated by looking for all_scores record
	_, err = s.getAllScores(ctx, gameID)
	if err == nil {
		// Already migrated
		return nil
	}

	// Create all scores record from existing leaderboard entries
	allScores := &models.AllScoresRecord{
		GameID:  gameID,
		Scores:  leaderboard.Entries,
		Updated: time.Now(),
	}

	// Save all scores
	var buf strings.Builder
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(allScores); err != nil {
		return fmt.Errorf("failed to marshal all scores during migration: %w", err)
	}
	jsonData := strings.TrimSuffix(buf.String(), "\n")
	if err := s.db.Set(ctx, fmt.Sprintf("all_scores:%s", gameID), jsonData); err != nil {
		return fmt.Errorf("failed to save all scores during migration: %w", err)
	}

	// Create player high scores from existing entries
	highScores := &models.PlayerHighScores{
		GameID:     gameID,
		HighScores: make(map[string]models.ScoreEntry),
		Updated:    time.Now(),
	}

	// For each entry, keep only the highest score per initials
	for _, entry := range leaderboard.Entries {
		existing, exists := highScores.HighScores[entry.Initials]
		if !exists || entry.Score > existing.Score {
			highScores.HighScores[entry.Initials] = entry
		}
	}

	// Save player high scores
	buf.Reset()
	if err := encoder.Encode(highScores); err != nil {
		return fmt.Errorf("failed to marshal high scores during migration: %w", err)
	}
	jsonData = strings.TrimSuffix(buf.String(), "\n")
	if err := s.db.Set(ctx, fmt.Sprintf("player_high_scores:%s", gameID), jsonData); err != nil {
		return fmt.Errorf("failed to save high scores during migration: %w", err)
	}

	// Regenerate the filtered leaderboard to ensure consistency
	return s.regenerateFilteredLeaderboard(ctx, gameID)
}
