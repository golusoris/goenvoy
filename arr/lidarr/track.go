package lidarr

import (
	"context"
	"fmt"
)

// GetTracks returns tracks for the given artist and album.
func (c *Client) GetTracks(ctx context.Context, artistID, albumID int) ([]Track, error) {
	var out []Track
	path := fmt.Sprintf("/api/v1/track?artistId=%d&albumId=%d", artistID, albumID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTrack returns a single track by its database ID.
func (c *Client) GetTrack(ctx context.Context, id int) (*Track, error) {
	var out Track
	path := fmt.Sprintf("/api/v1/track/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTrackFiles returns all track files for the given artist.
func (c *Client) GetTrackFiles(ctx context.Context, artistID int) ([]TrackFile, error) {
	var out []TrackFile
	path := fmt.Sprintf("/api/v1/trackfile?artistId=%d", artistID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTrackFile returns a single track file by its database ID.
func (c *Client) GetTrackFile(ctx context.Context, id int) (*TrackFile, error) {
	var out TrackFile
	path := fmt.Sprintf("/api/v1/trackfile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTrackFile removes a single track file by its database ID.
func (c *Client) DeleteTrackFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/trackfile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteTrackFiles removes multiple track files by their IDs.
func (c *Client) DeleteTrackFiles(ctx context.Context, ids []int) error {
	body := TrackFileListResource{TrackFileIDs: ids}
	return c.base.Delete(ctx, "/api/v1/trackfile/bulk", &body, nil)
}

// UpdateTrackFile updates a single track file.
func (c *Client) UpdateTrackFile(ctx context.Context, tf *TrackFile) (*TrackFile, error) {
	var out TrackFile
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v1/trackfile/%d", tf.ID), tf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditTrackFiles performs a bulk edit of track file properties.
func (c *Client) EditTrackFiles(ctx context.Context, editor *TrackFileEditorResource) ([]TrackFile, error) {
	var out []TrackFile
	if err := c.base.Put(ctx, "/api/v1/trackfile/editor", editor, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Album Studio ----------.
