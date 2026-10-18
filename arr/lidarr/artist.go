package lidarr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetAllArtists returns every artist configured in Lidarr.
func (c *Client) GetAllArtists(ctx context.Context) ([]Artist, error) {
	return arr.GetList[Artist](ctx, c.base, "/api/v1/artist")
}

// GetArtist returns a single artist by its database ID.
func (c *Client) GetArtist(ctx context.Context, id int) (*Artist, error) {
	path := fmt.Sprintf("/api/v1/artist/%d", id)
	return arr.GetOne[Artist](ctx, c.base, path)
}

// AddArtist adds a new artist to Lidarr.
func (c *Client) AddArtist(ctx context.Context, artist *Artist) (*Artist, error) {
	return arr.PostOne[Artist](ctx, c.base, "/api/v1/artist", artist)
}

// UpdateArtist updates an existing artist. Set moveFiles to true to relocate
// files when the artist path changes.
func (c *Client) UpdateArtist(ctx context.Context, artist *Artist, moveFiles bool) (*Artist, error) {
	path := fmt.Sprintf("/api/v1/artist/%d?moveFiles=%t", artist.ID, moveFiles)
	return arr.PutOne[Artist](ctx, c.base, path, artist)
}

// DeleteArtist removes an artist. Set deleteFiles to true to also delete
// downloaded files from disk.
func (c *Client) DeleteArtist(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/artist/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupArtist searches for an artist by name.
func (c *Client) LookupArtist(ctx context.Context, term string) ([]Artist, error) {
	path := "/api/v1/artist/lookup?term=" + url.QueryEscape(term)
	return arr.GetList[Artist](ctx, c.base, path)
}

// EditArtists performs a batch update on multiple artists.
func (c *Client) EditArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Put(ctx, "/api/v1/artist/editor", editor, nil)
}

// DeleteArtists performs a batch delete of multiple artists.
func (c *Client) DeleteArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/artist/editor", editor, nil)
}
