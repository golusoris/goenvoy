package radarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/radarr"
)

func Example() {
	// Create a new Radarr client
	client, err := radarr.New("http://localhost:7878", "your-api-key")
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

	// Get all movies
	movies, err := client.GetAllMovies(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total movies: %d\n", len(movies))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := radarr.New("http://localhost:7878", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, radarr.QueueOptions{
		Page:           1,
		PageSize:       20,
		IncludeUnknown: false,
		IncludeParent:  false, // ignored by Radarr
		IncludeItem:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Queue records: %d\n", len(queue.Records))
}

func ExampleClient_DeleteQueueItemWithOptions() {
	client, err := radarr.New("http://localhost:7878", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, radarr.QueueDeleteOptions{
		RemoveFromClient: true,
		Blocklist:        false,
		SkipRedownload:   true,
		ChangeCategory:   false,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_GetManualImportCandidates() {
	client, err := radarr.New("http://localhost:7878", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	candidates, err := client.GetManualImportCandidates(context.Background(), radarr.ManualImportOptions{
		DownloadID: "SABnzbd_nzo_abc123", FilterExistingFiles: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	for i := range candidates {
		fmt.Printf("%s: %d rejections\n", candidates[i].Path, len(candidates[i].Rejections))
	}
}

func ExampleClient_SendManualImport() {
	client, err := radarr.New("http://localhost:7878", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	candidates, err := client.GetManualImportCandidates(ctx, radarr.ManualImportOptions{
		DownloadID: "SABnzbd_nzo_abc123", FilterExistingFiles: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Import only candidates without rejections that are mapped to a movie.
	var files []radarr.ManualImportFile
	for i := range candidates {
		if f, ok := candidates[i].ImportFile(); ok && len(candidates[i].Rejections) == 0 {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return
	}

	cmd, err := client.SendManualImport(ctx, radarr.ImportModeAuto, files)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("command %d queued\n", cmd.ID)
}
