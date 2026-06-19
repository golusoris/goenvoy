package rtorrent_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/downloadclient/rtorrent/v2"
)

func Example() {
	// Create a new rTorrent client
	client, err := rtorrent.New("http://localhost:8000/RPC2")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get rTorrent system info
	info, err := client.GetSystemInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("rTorrent: %s\n", info.LibraryVersion)

	// Get download list
	torrents, err := client.GetTorrents(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total torrents: %d\n", len(torrents))
}
