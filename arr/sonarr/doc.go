// Package sonarr provides a client for the Sonarr v3 API.
//
// Sonarr is a PVR for Usenet and BitTorrent users that monitors
// multiple RSS feeds for new episodes of TV series and automatically
// grabs, sorts, and renames them. The v3 API applies to both v3 and
// v4 versions of the Sonarr application.
//
// The [Client] type wraps [arr.BaseClient] and exposes typed methods
// for every major Sonarr resource: series, episodes, episode files,
// calendar, queue, commands, history, and more.
package sonarr
