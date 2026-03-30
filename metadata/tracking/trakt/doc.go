// Package trakt provides a client for the Trakt API v2.
//
// Trakt (https://trakt.tv) is a platform for tracking movies, TV shows,
// and anime. The API provides access to movie/show metadata, trending lists,
// ratings, people, search, calendars, genres, certifications, and more.
//
// # Authentication
//
// All requests require a client ID passed via the trakt-api-key header.
// User-specific endpoints (sync, watchlist, etc.) additionally require OAuth2
// tokens, which are not covered by this client.
//
// # Usage
//
//	client := trakt.New("your-client-id")
//	movie, err := client.GetMovie(ctx, "the-dark-knight")
//	if err != nil {
//		log.Fatal(err)
//	}
package trakt
