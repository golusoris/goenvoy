package readarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/readarr"
)

func Example() {
	// Create a new Readarr client
	client, err := readarr.New("http://localhost:8787", "your-api-key")
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

	// Get all authors
	authors, err := client.GetAllAuthors(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total authors: %d\n", len(authors))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := readarr.New("http://localhost:8787", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, readarr.QueueOptions{
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
	client, err := readarr.New("http://localhost:8787", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, readarr.QueueDeleteOptions{
		RemoveFromClient: true,
		Blocklist:        false,
		SkipRedownload:   true,
		ChangeCategory:   false,
	})
	if err != nil {
		log.Fatal(err)
	}
}
