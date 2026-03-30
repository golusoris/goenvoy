// Package shoko provides a client for the Shoko Server REST API v3.
//
// Shoko Server (https://shokoanime.com) is an anime-focused media management
// server that organizes, identifies, and tags anime files using AniDB metadata.
// It also integrates with TMDB for supplementary data.
//
// # Authentication
//
// Authenticate by calling [Client.Login] with a username and password.
// The returned API key is stored on the client and sent as the "apikey" header
// with every subsequent request.
//
// # Usage
//
//	client := shoko.New("http://localhost:8111")
//	if err := client.Login(ctx, "admin", "password"); err != nil {
//		log.Fatal(err)
//	}
//	series, err := client.GetSeries(ctx, 1)
package shoko
