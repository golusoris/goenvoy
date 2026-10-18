package sonarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/sonarr"
)

func Example() {
	// Create a new Sonarr client
	client, err := sonarr.New("http://localhost:8989", "your-api-key")
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

	// Get all series
	series, err := client.GetAllSeries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total series: %d\n", len(series))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := sonarr.New("http://localhost:8989", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, sonarr.QueueOptions{
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
	client, err := sonarr.New("http://localhost:8989", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, sonarr.QueueDeleteOptions{
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
	client, err := sonarr.New("http://localhost:8989", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	candidates, err := client.GetManualImportCandidates(context.Background(), sonarr.ManualImportOptions{
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
	client, err := sonarr.New("http://localhost:8989", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	candidates, err := client.GetManualImportCandidates(ctx, sonarr.ManualImportOptions{
		DownloadID: "SABnzbd_nzo_abc123", FilterExistingFiles: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Import only candidates without rejections that are mapped to a series.
	var files []sonarr.ManualImportFile
	for i := range candidates {
		if f, ok := candidates[i].ImportFile(); ok && len(candidates[i].Rejections) == 0 {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return
	}

	cmd, err := client.SendManualImport(ctx, sonarr.ImportModeAuto, files)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("command %d queued\n", cmd.ID)
}
