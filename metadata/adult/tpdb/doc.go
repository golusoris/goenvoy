// Package tpdb provides a client for ThePornDB (TPDB) API.
//
// ThePornDB (https://theporndb.net) is a metadata database for adult content.
// The API provides access to scenes, performers, sites/studios, movies,
// JAV content, tags, and directors.
//
// # Authentication
//
// All requests require a Bearer token passed in the Authorization header.
// Obtain a token from https://theporndb.net/user/api-tokens.
//
// # Usage
//
//	client := tpdb.New("your-api-token")
//	scene, err := client.GetScene(ctx, "abc123")
//	if err != nil {
//		log.Fatal(err)
//	}
package tpdb
