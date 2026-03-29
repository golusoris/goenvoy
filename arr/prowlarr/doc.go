// Package prowlarr provides a client for the Prowlarr v1 API.
//
// Prowlarr is an indexer manager and proxy that integrates with various
// PVR applications. It manages all indexer configuration in a single
// place and pushes them to connected applications like Sonarr, Radarr,
// Lidarr, and Readarr.
//
// The [Client] type wraps [arr.BaseClient] and exposes typed methods
// for every major Prowlarr resource: indexers, applications, app
// profiles, search, history, statistics, tags, commands, and more.
package prowlarr
