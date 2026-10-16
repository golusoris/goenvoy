package queue_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/downloadclient/queue"
)

func ExampleTake() {
	target := queue.Target{
		Kind:     queue.KindQBittorrent,
		BaseURL:  "http://localhost:8080",
		Username: "admin",
		Secret:   "adminadmin",
	}

	snap, err := queue.Take(context.Background(), target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client: %s (version: %s)\n", snap.Kind, snap.Version)
	fmt.Printf("Active: %d, Queued: %d\n", snap.Active, snap.Queued)
	fmt.Printf("Download rate: %d B/s\n", snap.DownloadRate)
}
