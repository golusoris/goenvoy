// Package retroachievements provides a client for the RetroAchievements API.
//
// RetroAchievements is a community-driven project providing achievement tracking
// for retro games. Authentication is via a web API key passed as a query parameter.
//
// Usage:
//
//	c := retroachievements.New("your-api-key")
//	game, err := c.GetGame(context.Background(), 1)
package retroachievements
