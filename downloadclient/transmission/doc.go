// Package transmission provides a client for the Transmission RPC API.
//
// Transmission (https://transmissionbt.com) exposes a JSON-RPC interface
// for managing torrents. The RPC endpoint is typically at /transmission/rpc.
//
// # Authentication
//
// Transmission uses a CSRF protection scheme with a session ID header.
// The client automatically handles the X-Transmission-Session-Id negotiation:
// if a request returns HTTP 409, the client retries with the new session ID.
// Basic HTTP authentication is also supported when the server requires it.
//
// # Usage
//
//	client := transmission.New("http://localhost:9091")
//	torrents, err := client.GetTorrents(ctx, nil)
package transmission
