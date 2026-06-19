package tdarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/tdarr/v3"
)

func Example() {
	client, err := tdarr.New("http://localhost:8265", tdarr.WithAPIKey("your-key"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	status, err := client.GetStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tdarr %s on %s\n", status.Version, status.Os)
}
