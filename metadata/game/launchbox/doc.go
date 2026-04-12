// Package launchbox provides a client for the LaunchBox Games Database.
//
// LaunchBox distributes its games database as a downloadable Metadata.zip file
// containing XML data for games, platforms, alternate names, and images. This
// client downloads and parses that database, providing in-memory search and
// lookup methods.
//
// Usage:
//
//	c := launchbox.New()
//	if err := c.Download(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
//	games := c.SearchGames("Mario", "Nintendo Entertainment System")
package launchbox
