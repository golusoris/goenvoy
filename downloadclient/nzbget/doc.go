// Package nzbget provides a client for the NZBGet JSON-RPC API.
//
// NZBGet (https://nzbget.com) is a Usenet download client that exposes a
// JSON-RPC API at /jsonrpc. Authentication uses HTTP basic auth.
//
// # Authentication
//
// The client uses HTTP basic authentication with the username and password
// configured in NZBGet.
//
// # Usage
//
//	client := nzbget.New("http://localhost:6789", "nzbget", "password")
//	groups, err := client.ListGroups(ctx)
package nzbget
