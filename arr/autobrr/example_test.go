package autobrr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/autobrr/v2"
)

func Example() {
	client, err := autobrr.New("http://localhost:7474", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get all filters
	filters, err := client.GetFilters(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Filters: %d\n", len(filters))

	// Get IRC networks
	networks, err := client.GetIRCNetworks(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("IRC Networks: %d\n", len(networks))
}
