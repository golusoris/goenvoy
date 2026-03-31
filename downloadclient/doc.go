// Package downloadclient provides shared types and interfaces for interacting
// with download client APIs (qBittorrent, Transmission, Deluge, rTorrent,
// SABnzbd, NZBGet).
//
// Individual client packages implement the [Downloader] interface for their
// respective applications, allowing consumers to swap download clients without
// changing business logic.
package downloadclient
