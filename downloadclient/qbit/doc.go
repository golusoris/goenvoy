// Package qbit provides a client for the qBittorrent WebUI API (v2).
//
// qBittorrent (https://www.qbittorrent.org) exposes a REST-like WebUI API
// for managing torrents, transfer settings, categories, and tags.
//
// # Authentication
//
// The client authenticates via cookie-based sessions. Call [Client.Login] to
// obtain a session, then use the client for subsequent requests. The SID cookie
// is automatically managed by the underlying [http.Client] cookie jar.
//
// # Usage
//
//	client := qbit.New("http://localhost:8080")
//	if err := client.Login(ctx, "admin", "adminadmin"); err != nil {
//		log.Fatal(err)
//	}
//	torrents, err := client.ListTorrents(ctx, nil)
package qbit
