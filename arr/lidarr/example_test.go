package lidarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/lidarr"
)

func Example() {
	// Create a new Lidarr client
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get system status
	status, err := client.GetSystemStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s version %s\n", status.AppName, status.Version)

	// Get all artists
	artists, err := client.GetAllArtists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total artists: %d\n", len(artists))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, lidarr.QueueOptions{
		Page:           1,
		PageSize:       20,
		IncludeUnknown: false,
		IncludeParent:  true,
		IncludeItem:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue records: %d\n", len(queue.Records))
}

func ExampleClient_DeleteQueueItemWithOptions() {
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, lidarr.QueueDeleteOptions{
		RemoveFromClient: true,
		Blocklist:        false,
		SkipRedownload:   true,
		ChangeCategory:   false,
	})
	if err != nil {
		log.Fatal(err)
	}
}
