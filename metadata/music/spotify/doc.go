// Package spotify provides a client for the Spotify Web API.
//
// Spotify provides music metadata including artists, albums, tracks, audio
// features, recommendations, and browse categories.
//
// GetArtistTopTracks, GetNewReleases, and GetCategories require extended quota
// access for Development Mode Client IDs created on or after February 11, 2026.
// Spotify postponed this change for existing Development Mode integrations.
// GetRelatedArtists, GetAudioFeatures, and GetRecommendations are not listed
// as available for Development Mode apps and require extended quota access.
//
// Usage:
//
//	c := spotify.New("your-access-token")
//	artist, err := c.GetArtist(context.Background(), "0OdUWJ0sBjDrqHygGUXeCF")
package spotify
