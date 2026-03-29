// Package arr provides shared types and a base HTTP client for interacting
// with *arr application APIs (Sonarr, Radarr, Lidarr, Readarr, Prowlarr,
// Bazarr, Whisparr, Seerr).
//
// The [BaseClient] handles authentication, request construction, and response
// parsing common to all *arr services. Individual service packages build on
// top of this foundation.
//
// # Authentication
//
// Every request is authenticated via the X-Api-Key header. Pass the API key
// when constructing a [BaseClient]:
//
//	c, err := arr.NewBaseClient("http://localhost:8989", "your-api-key")
//
// # Error Handling
//
// Non-2xx responses are returned as [*APIError], which exposes the HTTP status
// code, request method, path, and response body for inspection.
package arr
