// Package spotify provides a client for the Spotify Web API.
//
// Spotify provides music metadata including artists, albums, tracks, audio
// features, recommendations, and browse categories.
//
// GetArtistTopTracks, GetNewReleases, GetCategories, GetRelatedArtists,
// GetAudioFeatures, and GetRecommendations require extended quota access
// following the February 2026 Development Mode changes.
//
// Usage:
//
//	c := spotify.New("your-access-token")
//	artist, err := c.GetArtist(context.Background(), "0OdUWJ0sBjDrqHygGUXeCF")
package spotify
