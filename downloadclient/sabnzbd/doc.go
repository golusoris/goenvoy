// Package sabnzbd provides a client for the SABnzbd API.
//
// SABnzbd (https://sabnzbd.org) is a Usenet download client that exposes a
// REST API for managing NZB downloads. All requests require an API key passed
// as a query parameter.
//
// # Authentication
//
// Authentication uses an API key passed with every request as a query parameter.
//
// # Usage
//
//	client := sabnzbd.New("http://localhost:8080", "your-api-key")
//	queue, err := client.GetQueue(ctx)
package sabnzbd
