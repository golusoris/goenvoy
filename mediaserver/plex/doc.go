// Package plex provides a client for the Plex Media Server API.
//
// Plex (https://plex.tv) is a media server that organizes and streams your
// personal media collection. The API provides access to libraries, metadata,
// playback sessions, search, and server management.
//
// # Authentication
//
// All requests require an X-Plex-Token. You can obtain one from your
// Plex account settings or by logging in via [Client.SignIn].
//
// # Usage
//
//	client := plex.New("http://192.168.1.100:32400", "your-token")
//	libs, err := client.GetLibraries(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
package plex
