package whisparr_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/golusoris/goenvoy/arr/whisparr"
)

func Example() {
	// Create a new Whisparr v2 client (Sonarr-based).
	client, err := whisparr.New("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get system status.
	status, err := client.GetSystemStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s version %s\n", status.AppName, status.Version)

	// Get all series.
	series, err := client.GetAllSeries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total series: %d\n", len(series))
}

func Example_v3() {
	// Create a new Whisparr v3 client (Radarr-based).
	client, err := whisparr.NewV3("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get all movies.
	movies, err := client.GetAllMovies(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total movies: %d\n", len(movies))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := whisparr.New("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, whisparr.QueueOptions{
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
	client, err := whisparr.New("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, whisparr.QueueDeleteOptions{
		RemoveFromClient: true,
		Blocklist:        false,
		SkipRedownload:   true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClientV3_GetQueueWithOptions() {
	client, err := whisparr.NewV3("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, whisparr.QueueOptions{
		Page:           1,
		PageSize:       20,
		IncludeUnknown: false,
		IncludeParent:  false, // ignored by Whisparr v3
		IncludeItem:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue records: %d\n", len(queue.Records))
}

func ExampleClientV3_DeleteQueueItemWithOptions() {
	client, err := whisparr.NewV3("http://localhost:6969", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, whisparr.QueueDeleteOptions{
		RemoveFromClient: true,
		Blocklist:        false,
		SkipRedownload:   true,
		ChangeCategory:   false,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClientV3_GetManualImportCandidates() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[{"id":1,"name":"scene.mkv","relativePath":"scene.mkv","path":"/downloads/scene.mkv","size":1048576,"quality":{"quality":{"id":1,"name":"SDTV","source":"television","resolution":480},"revision":{"version":1,"real":0,"isRepack":false}},"languages":[],"customFormats":[],"rejections":[]}]`)
	}))
	defer server.Close()

	client, err := whisparr.NewV3(server.URL, "test-key")
	if err != nil {
		return
	}

	candidates, err := client.GetManualImportCandidates(context.Background(), whisparr.ManualImportOptionsV3{
		DownloadID:          "ABC123",
		FilterExistingFiles: true,
	})
	if err != nil {
		return
	}

	for _, c := range candidates {
		fmt.Println(c.Name)
	}
	// Output:
	// scene.mkv
}
