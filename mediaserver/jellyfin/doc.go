// Package jellyfin provides a client for the Jellyfin Media Server API.
//
// Jellyfin (https://jellyfin.org) is an open-source media server that organizes
// and streams your personal media collection. The API provides access to libraries,
// metadata, playback sessions, search, and server management.
//
// # Authentication
//
// Authenticate by calling [Client.AuthenticateByName] to obtain an access token,
// which is then sent as a Bearer token with every request.
//
// # Usage
//
//	client := jellyfin.New("http://192.168.1.100:8096")
//	if err := client.AuthenticateByName(ctx, "username", "password"); err != nil {
//		log.Fatal(err)
//	}
//	items, err := client.GetItems(ctx, nil)
package jellyfin
