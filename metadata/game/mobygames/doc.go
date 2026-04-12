// Package mobygames provides a client for the MobyGames API.
//
// MobyGames is a comprehensive database of video games. Authentication is via
// an API key passed as a query parameter.
//
// Usage:
//
//	c := mobygames.New("your-api-key")
//	result, err := c.SearchGames(context.Background(), "zelda", 0, 10)
package mobygames
