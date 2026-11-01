package sabnzbd_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/golusoris/goenvoy/downloadclient/sabnzbd/v2"
)

func Example() {
	// Create a new SABnzbd client
	client, err := sabnzbd.New("http://localhost:8080", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get version
	version, err := client.GetVersion(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SABnzbd version: %s\n", version)

	// Get queue
	queue, err := client.GetQueue(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queue slots: %d\n", len(queue.Slots))
}

func ExampleClient_SetPriority() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"position": 1}`))
	}))
	defer ts.Close()

	client, err := sabnzbd.New(ts.URL, "test-key")
	if err != nil {
		return
	}

	pos, err := client.SetPriority(context.Background(), "SABnzbd_nzo_abc", sabnzbd.PriorityHigh)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("New queue position: %d\n", pos)

	// Output:
	// New queue position: 1
}
