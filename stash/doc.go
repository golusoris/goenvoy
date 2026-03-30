// Package stash provides a client for the Stash GraphQL API.
//
// Stash (https://github.com/stashapp/stash) is a self-hosted web application
// for organizing and serving adult media. The API provides access to scenes,
// performers, studios, tags, galleries, images, groups, and scene markers.
//
// # Authentication
//
// Requests are authenticated via the ApiKey header:
//
//	client := stash.New("http://localhost:9999/graphql", "your-api-key")
//	scene, err := client.FindScene(ctx, "1")
//	if err != nil {
//		log.Fatal(err)
//	}
package stash
