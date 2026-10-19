package lidarr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/arr/lidarr"
)

func Example() {
	// Create a new Lidarr client
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
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

	// Get all artists
	artists, err := client.GetAllArtists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total artists: %d\n", len(artists))
}

func ExampleClient_GetQueueWithOptions() {
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	queue, err := client.GetQueueWithOptions(ctx, lidarr.QueueOptions{
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
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.DeleteQueueItemWithOptions(ctx, 42, lidarr.QueueDeleteOptions{
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
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	candidates, err := client.GetManualImportCandidates(context.Background(), lidarr.ManualImportOptions{
		DownloadID: "SABnzbd_nzo_abc123", FilterExistingFiles: true, ReplaceExistingFiles: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	for i := range candidates {
		fmt.Printf("%s: %d rejections\n", candidates[i].Path, len(candidates[i].Rejections))
	}
}

func ExampleClient_SendManualImport() {
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	candidates, err := client.GetManualImportCandidates(ctx, lidarr.ManualImportOptions{
		DownloadID: "SABnzbd_nzo_abc123", FilterExistingFiles: true, ReplaceExistingFiles: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Import only candidates without rejections that are mapped to an artist and album.
	var files []lidarr.ManualImportFile
	for i := range candidates {
		if f, ok := candidates[i].ImportFile(); ok && len(candidates[i].Rejections) == 0 {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return
	}

	cmd, err := client.SendManualImport(ctx, lidarr.ImportModeAuto, files)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("command %d queued\n", cmd.ID)
}

func ExampleClient_RescoreManualImport() {
	client, err := lidarr.New("http://localhost:8686", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	candidates, err := client.RescoreManualImport(ctx, []lidarr.ManualImportUpdate{
		{
			Path:                    "/downloads/complete/Album/01.mp3",
			AlbumID:                 34,
			AlbumReleaseID:          56,
			DisableReleaseSwitching: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := range candidates {
		fmt.Printf("%s: %d rejections\n", candidates[i].Path, len(candidates[i].Rejections))
	}
}
