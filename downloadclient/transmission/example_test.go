package transmission_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/downloadclient/transmission/v2"
)

func Example() {
	// Create a new Transmission client
	client, err := transmission.New("http://localhost:9091/transmission/rpc",
		transmission.WithAuth("username", "password"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get session stats
	stats, err := client.GetSessionStats(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Active torrents: %d\n", stats.ActiveTorrentCount)

	// List all torrents
	torrents, err := client.GetTorrents(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total torrents: %d\n", len(torrents))
}
