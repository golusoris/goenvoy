package plex_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/plex/v2"
)

func Example() {
	// Create a new Plex client
	client, err := plex.New("http://192.168.1.100:32400", "your-plex-token")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get server identity (no authentication required)
	identity, err := client.GetIdentity(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server: %s\n", identity.Version)

	// Get libraries
	libraries, err := client.GetLibraries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Libraries: %d\n", len(libraries))
}
