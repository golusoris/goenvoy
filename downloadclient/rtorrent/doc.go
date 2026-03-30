// Package rtorrent provides a client for the rTorrent XML-RPC API.
//
// rTorrent (https://rakshasa.github.io/rtorrent/) exposes an XML-RPC
// interface for managing torrents. The client communicates via HTTP POST
// with XML-RPC encoded requests to the RPC endpoint.
//
// # Authentication
//
// Authentication is handled via HTTP basic auth when the rTorrent SCGI
// proxy (commonly nginx or Apache) is configured to require it.
//
// # Usage
//
//	client := rtorrent.New("http://localhost:8080/RPC2")
//	torrents, err := client.GetTorrents(ctx)
package rtorrent
