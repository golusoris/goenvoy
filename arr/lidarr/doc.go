// Package lidarr provides a client for the Lidarr v1 API.
//
// Lidarr is a music collection manager for Usenet and BitTorrent users
// that monitors multiple RSS feeds for new albums and automatically
// grabs, sorts, and renames them. The v1 API applies to current
// versions of the Lidarr application.
//
// The [Client] type wraps [arr.BaseClient] and exposes typed methods
// for every major Lidarr resource: artists, albums, tracks, track files,
// calendar, queue, commands, history, and more.
package lidarr
