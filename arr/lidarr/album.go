package lidarr

import (
	"context"
	"fmt"
	"net/url"
)

// GetAlbums returns albums for the given artist.
func (c *Client) GetAlbums(ctx context.Context, artistID int) ([]Album, error) {
	var out []Album
	path := fmt.Sprintf("/api/v1/album?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAlbum returns a single album by its database ID.
func (c *Client) GetAlbum(ctx context.Context, id int) (*Album, error) {
	var out Album
	path := fmt.Sprintf("/api/v1/album/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddAlbum adds a new album to Lidarr.
func (c *Client) AddAlbum(ctx context.Context, album *Album) (*Album, error) {
	var out Album
	if err := c.base.Post(ctx, "/api/v1/album", album, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAlbum updates an existing album.
func (c *Client) UpdateAlbum(ctx context.Context, album *Album) (*Album, error) {
	var out Album
	path := fmt.Sprintf("/api/v1/album/%d", album.ID)
	if err := c.base.Put(ctx, path, album, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAlbum removes an album.
func (c *Client) DeleteAlbum(ctx context.Context, id int, deleteFiles, addImportListExclusion bool) error {
	path := fmt.Sprintf("/api/v1/album/%d?deleteFiles=%t&addImportListExclusion=%t", id, deleteFiles, addImportListExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupAlbum searches for an album by name.
func (c *Client) LookupAlbum(ctx context.Context, term string) ([]Album, error) {
	var out []Album
	path := "/api/v1/album/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorAlbums sets the monitored status for a list of albums.
func (c *Client) MonitorAlbums(ctx context.Context, req *AlbumsMonitoredResource) error {
	return c.base.Put(ctx, "/api/v1/album/monitor", req, nil)
}

// AlbumStudio performs batch monitoring changes on artists and albums.
func (c *Client) AlbumStudio(ctx context.Context, studio *AlbumStudioResource) error {
	return c.base.Post(ctx, "/api/v1/albumstudio", studio, nil)
}

// ---------- Calendar By ID ----------.
