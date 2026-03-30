// Package simkl provides a client for the Simkl API.
//
// Simkl (https://simkl.com) is a platform for tracking movies, TV shows,
// and anime. The API provides access to media metadata, trending lists,
// genre filtering, premieres, airing schedules, search, and calendars.
//
// # Authentication
//
// All requests require a client ID passed via the simkl-api-key header.
// User-specific endpoints (sync, ratings, etc.) additionally require OAuth2
// tokens, which are not covered by this client.
//
// # Usage
//
//	client := simkl.New("your-client-id")
//	movie, err := client.GetMovie(ctx, "the-dark-knight")
//	if err != nil {
//		log.Fatal(err)
//	}
package simkl
