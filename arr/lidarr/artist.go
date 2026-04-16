package lidarr

import (
	"context"
	"fmt"
	"net/url"
)

// GetAllArtists returns every artist configured in Lidarr.
func (c *Client) GetAllArtists(ctx context.Context) ([]Artist, error) {
	var out []Artist
	if err := c.base.Get(ctx, "/api/v1/artist", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetArtist returns a single artist by its database ID.
func (c *Client) GetArtist(ctx context.Context, id int) (*Artist, error) {
	var out Artist
	path := fmt.Sprintf("/api/v1/artist/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddArtist adds a new artist to Lidarr.
func (c *Client) AddArtist(ctx context.Context, artist *Artist) (*Artist, error) {
	var out Artist
	if err := c.base.Post(ctx, "/api/v1/artist", artist, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateArtist updates an existing artist. Set moveFiles to true to relocate
// files when the artist path changes.
func (c *Client) UpdateArtist(ctx context.Context, artist *Artist, moveFiles bool) (*Artist, error) {
	var out Artist
	path := fmt.Sprintf("/api/v1/artist/%d?moveFiles=%t", artist.ID, moveFiles)
	if err := c.base.Put(ctx, path, artist, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteArtist removes an artist. Set deleteFiles to true to also delete
// downloaded files from disk.
func (c *Client) DeleteArtist(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/artist/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupArtist searches for an artist by name.
func (c *Client) LookupArtist(ctx context.Context, term string) ([]Artist, error) {
	var out []Artist
	path := "/api/v1/artist/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EditArtists performs a batch update on multiple artists.
func (c *Client) EditArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Put(ctx, "/api/v1/artist/editor", editor, nil)
}

// DeleteArtists performs a batch delete of multiple artists.
func (c *Client) DeleteArtists(ctx context.Context, editor *ArtistEditorResource) error {
	return c.base.Delete(ctx, "/api/v1/artist/editor", editor, nil)
}
