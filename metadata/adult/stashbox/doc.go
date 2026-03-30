// Package stashbox provides a client for the StashBox GraphQL API.
//
// StashBox (https://github.com/stashapp/stash-box) is an open-source metadata
// repository for adult content cataloging. It provides a GraphQL API for
// querying performers, scenes, studios, tags, and sites.
//
// # Authentication
//
// All requests require an API key passed via the ApiKey header:
//
//	client := stashbox.New("https://stashdb.org/graphql", "your-api-key")
//	performer, err := client.FindPerformer(ctx, "uuid-here")
//	if err != nil {
//		log.Fatal(err)
//	}
package stashbox
