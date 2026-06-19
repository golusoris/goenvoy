package nzbhydra_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/nzbhydra/v2"
)

func Example() {
	client, err := nzbhydra.New("http://localhost:5076", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Search all indexers
	results, err := client.Search(ctx, "ubuntu", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Results: %d\n", len(results))
}
