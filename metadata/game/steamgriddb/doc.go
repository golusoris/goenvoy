// Package steamgriddb provides a client for the SteamGridDB API.
//
// SteamGridDB provides custom artwork (grids, heroes, logos, icons) for games.
// Authentication is via a Bearer token in the Authorization header.
//
// Usage:
//
//	c := steamgriddb.New("your-api-key")
//	game, err := c.GetGameByID(context.Background(), 12345)
package steamgriddb
