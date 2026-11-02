package launchbox

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/golusoris/goenvoy/metadata"
)

const defaultMetadataURL = "https://gamesdb.launchbox-app.com/Metadata.zip"

// Client is a LaunchBox Games Database client. Call [Client.Download] to fetch
// and parse the database before using lookup methods.
type Client struct {
	*metadata.BaseClient

	mu             sync.RWMutex
	games          []Game
	alternateNames []GameAlternateName
	images         []GameImage
	platforms      []Platform
	gamesByID      map[int]*Game
}

// APIError is returned when the download fails with a non-2xx status.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("launchbox: %s: %s", e.Status, e.Body)
}

// New creates a LaunchBox [Client].
func New(opts ...metadata.Option) *Client {
	bc := metadata.NewBaseClient(defaultMetadataURL, "launchbox", opts...)
	return &Client{
		BaseClient: bc,
		gamesByID:  make(map[int]*Game),
	}
}

// Download fetches the LaunchBox Metadata.zip and parses the XML contents.
// This must be called before using any lookup methods.
// A failure to close the response body is returned joined with any other error.
func (c *Client) Download(ctx context.Context) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL(), http.NoBody)
	if err != nil {
		return fmt.Errorf("launchbox: create request: %w", err)
	}

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("launchbox: request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("launchbox: close response body: %w", cerr))
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: string(body)}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("launchbox: read response: %w", err)
	}

	return c.parseZip(data)
}

// decodeEntry opens an archive entry, decodes XML into v, closes the reader, and joins any close error.
func decodeEntry(open func() (io.ReadCloser, error), name string, v any) (err error) {
	rc, err := open()
	if err != nil {
		return fmt.Errorf("launchbox: open %s: %w", name, err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("launchbox: close %s: %w", name, cerr))
		}
	}()

	if err := xml.NewDecoder(rc).Decode(v); err != nil {
		return fmt.Errorf("launchbox: parse %s: %w", name, err)
	}
	return nil
}

// parseEntries decodes recognized XML files from the archive into md and pd.
func parseEntries(files []*zip.File, md *metadataXML, pd *platformsXML) error {
	for _, f := range files {
		switch f.Name {
		case "Metadata.xml":
			if err := decodeEntry(f.Open, "Metadata.xml", md); err != nil {
				return err
			}
		case "Platforms.xml":
			if err := decodeEntry(f.Open, "Platforms.xml", pd); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseZip reads and decodes archive entries into local buffers before atomically publishing them.
func (c *Client) parseZip(data []byte) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("launchbox: open zip: %w", err)
	}

	var (
		md metadataXML
		pd platformsXML
	)
	if err := parseEntries(r.File, &md, &pd); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.games = md.Games
	c.alternateNames = md.AlternateNames
	c.images = md.Images
	c.platforms = pd.Platforms

	// Build index.
	c.gamesByID = make(map[int]*Game, len(c.games))
	for i := range c.games {
		c.gamesByID[c.games[i].DatabaseID] = &c.games[i]
	}

	return nil
}

// GetGameByID returns a game by its LaunchBox database ID.
func (c *Client) GetGameByID(id int) *Game {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gamesByID[id]
}

// SearchGames searches for games by name, optionally filtered by platform.
// The search is case-insensitive and matches substrings.
func (c *Client) SearchGames(query, platform string) []Game {
	c.mu.RLock()
	defer c.mu.RUnlock()

	q := strings.ToLower(query)
	p := strings.ToLower(platform)

	var results []Game
	for i := range c.games {
		g := &c.games[i]
		if !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		if p != "" && !strings.EqualFold(g.Platform, p) {
			continue
		}
		results = append(results, *g)
	}
	return results
}

// GetAlternateNames returns alternate names for a game by database ID.
func (c *Client) GetAlternateNames(databaseID int) []GameAlternateName {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []GameAlternateName
	for _, an := range c.alternateNames {
		if an.DatabaseID == databaseID {
			results = append(results, an)
		}
	}
	return results
}

// GetImages returns images for a game by database ID.
func (c *Client) GetImages(databaseID int) []GameImage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []GameImage
	for _, img := range c.images {
		if img.DatabaseID == databaseID {
			results = append(results, img)
		}
	}
	return results
}

// GetPlatforms returns all platforms in the database.
func (c *Client) GetPlatforms() []Platform {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]Platform, len(c.platforms))
	copy(out, c.platforms)
	return out
}

// GameCount returns the total number of games loaded.
func (c *Client) GameCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.games)
}
