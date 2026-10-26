package audiobookshelf_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/audiobookshelf/v2"
)

func Example() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get all libraries
	libraries, err := client.GetLibraries(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Libraries: %d\n", len(libraries))

	// Get server status
	status, err := client.GetStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Version: %s\n", status.ServerVersion)
}

func ExampleClient_Login() {
	client, err := audiobookshelf.New("http://localhost:13378", "")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Login(context.Background(), "admin", "password")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %s\n", resp.User.Username)
}

func ExampleClient_Logout() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Logout(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_GetUser() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	user, err := client.GetUser(context.Background(), "usr_1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %s\n", user.Username)
}

func ExampleClient_GetLibraryItemsWithOptions() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	items, err := client.GetLibraryItemsWithOptions(context.Background(), "lib_1", &audiobookshelf.LibraryItemsOptions{
		Page:     0,
		Limit:    25,
		Sort:     "media.metadata.title",
		Minified: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total: %d\n", items.Total)
}

func ExampleClient_GetPodcastFeed() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	feed, err := client.GetPodcastFeed(context.Background(), "https://example.com/feed.xml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Podcast: %s\n", feed.Metadata.Title)
}

func ExampleClient_CreatePodcast() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	item, err := client.CreatePodcast(context.Background(), &audiobookshelf.CreatePodcastRequest{
		LibraryID: "lib_1",
		FolderID:  "fol_1",
		Path:      "/podcasts/show",
		Media: audiobookshelf.CreatePodcastMedia{
			Metadata: audiobookshelf.PodcastMetadata{
				Title:   "Example Show",
				FeedURL: "https://example.com/feed.xml",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created item: %s\n", item.ID)
}

func ExampleClient_CheckNewEpisodes() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	episodes, err := client.CheckNewEpisodes(context.Background(), "li_pod_1", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New episodes: %d\n", len(episodes))
}

func ExampleClient_DownloadEpisodes() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	err = client.DownloadEpisodes(context.Background(), "li_pod_1", []audiobookshelf.PodcastFeedEpisode{
		{Title: "Episode 1"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_SearchPodcasts() {
	client, err := audiobookshelf.New("http://localhost:13378", "your-token")
	if err != nil {
		log.Fatal(err)
	}

	results, err := client.SearchPodcasts(context.Background(), "golang", "us")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Results: %d\n", len(results))
}

func ExampleClient_GetStatus() {
	client, err := audiobookshelf.New("http://localhost:13378", "")
	if err != nil {
		log.Fatal(err)
	}

	status, err := client.GetStatus(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initialized: %v\n", status.IsInit)
}
