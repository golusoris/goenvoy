package tpdb_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/adult/tpdb"
)

func Example() {
	// Create a new ThePornDB client
	client := tpdb.New("your-bearer-token")

	ctx := context.Background()

	// Search for performers
	results, _, err := client.SearchPerformers(ctx, &tpdb.PerformerSearchParams{Query: "Jane Doe"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d performers\n", len(results))
}

func ExampleClient_ListChanges() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":[{"id":101,"uuid":"00000000-0000-4000-8000-000000000001"}],"meta":{"total":1,"per_page":25,"current_page":1,"last_page":1}}`)
	}))
	defer ts.Close()

	client := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	records, _, err := client.ListChanges(context.Background(), tpdb.ChangeScenes, time.Time{}, 1, 25)
	if err != nil {
		return
	}
	fmt.Println(records[0].UUID)
	// Output:
	// 00000000-0000-4000-8000-000000000001
}

func ExampleClient_SearchPerformerSites() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":[{"id":"00000000-0000-4000-8000-000000000001","_id":1,"site_id":10,"name":"Fixture Performer","bio":"Bio","is_parent":true}],"meta":{"total":1,"per_page":25,"current_page":1,"last_page":1}}`)
	}))
	defer ts.Close()

	client := tpdb.New("test-token", metadata.WithBaseURL(ts.URL))
	sites, _, err := client.SearchPerformerSites(context.Background(), &tpdb.PerformerSiteSearchParams{Query: "Fixture"})
	if err != nil {
		return
	}
	fmt.Println(sites[0].Name)
	// Output:
	// Fixture Performer
}
