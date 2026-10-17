package seerr_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/seerr"
)

func Example() {
	// Create a new Seerr client
	client, err := seerr.New("http://localhost:5055", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get status
	status, err := client.GetStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Seerr version %s\n", status.Version)

	// Get requests
	requests, _, err := client.GetRequests(ctx, 25, 0, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total requests: %d\n", len(requests))
}

func ExampleMediaRequest_ServerAssigned() {
	// Seerr's first server has id 0; a request without a server carries null.
	for _, body := range []string{`{"id":1,"serverId":0}`, `{"id":2,"serverId":null}`} {
		var r seerr.MediaRequest
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			log.Fatal(err)
		}
		if r.ServerAssigned {
			fmt.Printf("request %d goes to server %d\n", r.ID, r.ServerID)
		} else {
			fmt.Printf("request %d uses the default server\n", r.ID)
		}
	}
	// Output:
	// request 1 goes to server 0
	// request 2 uses the default server
}
