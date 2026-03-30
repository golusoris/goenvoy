// Package deluge provides a client for the Deluge JSON-RPC API.
//
// Deluge (https://deluge-torrent.org) exposes a JSON-RPC interface on its
// web UI at /json. The client handles cookie-based session authentication
// automatically.
//
// # Authentication
//
// The client authenticates via the auth.login RPC method, which sets a session
// cookie. All subsequent requests reuse that cookie.
//
// # Usage
//
//	client := deluge.New("http://localhost:8112")
//	if err := client.Login(ctx, "deluge"); err != nil {
//		log.Fatal(err)
//	}
//	torrents, err := client.GetTorrentsStatus(ctx, nil)
package deluge
