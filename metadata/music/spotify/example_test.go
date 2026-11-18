package spotify_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/music/spotify"
)

func Example() {
	c := spotify.New("your-access-token")

	artist, err := c.GetArtist(context.Background(), "4gzpq5DPGxSnKTe4SA8HAU")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(artist.Name)
}

// ExampleClient_SetMaxResponseBytes configures the response body cap for Spotify requests.
func ExampleClient_SetMaxResponseBytes() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.Copy(w, strings.NewReader(`{"id":"1","name":"Spotify Artist"}`))
	}))
	defer srv.Close()

	c := spotify.New("token", metadata.WithBaseURL(srv.URL))
	c.SetMaxResponseBytes(1 << 20)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	artist, err := c.GetArtist(ctx, "1")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(artist.Name)
	// Output: Spotify Artist
}
