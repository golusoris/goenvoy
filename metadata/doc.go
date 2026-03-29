// Package metadata provides shared types for interacting with media metadata
// provider APIs (TMDb, TheTVDB, Fanart.tv, OMDb, TVmaze, AniList, Kitsu,
// AniDB, MyAnimeList, Trakt, Simkl, Letterboxd).
//
// Individual provider packages build on these shared types to offer
// provider-specific clients.
//
// # Shared Types
//
// [Rating], [ExternalID], [Image], [Person], and [SearchResult] are common
// across most metadata providers and allow consumers to work with normalised
// data regardless of the source.
package metadata
