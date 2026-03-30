// Package emby provides a client for the Emby Media Server API.
//
// Emby (https://emby.media) is a media server that organizes and streams
// your personal media collection. The API provides access to libraries,
// metadata, playback sessions, search, and server management.
//
// # Authentication
//
// Authenticate by calling [Client.AuthenticateByName] to obtain an access token,
// which is then sent with every request via the X-Emby-Token header.
//
// # Usage
//
//	client := emby.New("http://192.168.1.100:8096")
//	if err := client.AuthenticateByName(ctx, "username", "password"); err != nil {
//		log.Fatal(err)
//	}
//	items, err := client.GetItems(ctx, nil)
package emby
