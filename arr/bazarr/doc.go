// Package bazarr provides a client for the Bazarr API.
//
// Bazarr is a companion application that manages and downloads subtitles
// for media tracked by Sonarr and Radarr. It provides automatic subtitle
// searching, downloading, and upgrading.
//
// The [Client] type wraps [arr.BaseClient] and exposes typed methods for
// every major Bazarr resource: series, episodes, movies, subtitles,
// wanted items, history, providers, system status, and more.
package bazarr
