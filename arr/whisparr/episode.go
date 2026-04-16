package whisparr

import (
	"context"
	"fmt"
)

// GetEpisodes returns all episodes for a series.
func (c *Client) GetEpisodes(ctx context.Context, seriesID int) ([]Episode, error) {
	var out []Episode
	path := fmt.Sprintf("/api/v3/episode?seriesId=%d", seriesID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEpisode returns a single episode by ID.
func (c *Client) GetEpisode(ctx context.Context, id int) (*Episode, error) {
	var out Episode
	path := fmt.Sprintf("/api/v3/episode/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEpisodeFiles returns all episode files for a series.
func (c *Client) GetEpisodeFiles(ctx context.Context, seriesID int) ([]EpisodeFile, error) {
	var out []EpisodeFile
	path := fmt.Sprintf("/api/v3/episodefile?seriesId=%d", seriesID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteEpisodeFile deletes an episode file by ID.
func (c *Client) DeleteEpisodeFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v3/episodefile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// UpdateEpisode updates an episode (e.g. monitored status).
func (c *Client) UpdateEpisode(ctx context.Context, ep *Episode) (*Episode, error) {
	var out Episode
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/episode/%d", ep.ID), ep, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MonitorEpisodes sets the monitored status for multiple episodes.
func (c *Client) MonitorEpisodes(ctx context.Context, req *EpisodesMonitoredResource) ([]Episode, error) {
	var out []Episode
	if err := c.base.Put(ctx, "/api/v3/episode/monitor", req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Episode Files Extended ----------.

// GetEpisodeFile returns a single episode file by ID.
func (c *Client) GetEpisodeFile(ctx context.Context, id int) (*EpisodeFile, error) {
	var out EpisodeFile
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/episodefile/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEpisodeFile updates an episode file (quality, language).
func (c *Client) UpdateEpisodeFile(ctx context.Context, ef *EpisodeFile) (*EpisodeFile, error) {
	var out EpisodeFile
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/episodefile/%d", ef.ID), ef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditEpisodeFiles bulk edits episode files.
func (c *Client) EditEpisodeFiles(ctx context.Context, editor *EpisodeFileEditorResource) error {
	return c.base.Put(ctx, "/api/v3/episodefile/editor", editor, nil)
}

// BulkDeleteEpisodeFiles deletes multiple episode files.
func (c *Client) BulkDeleteEpisodeFiles(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/episodefile/bulk", struct {
		EpisodeFileIDs []int `json:"episodeFileIds"`
	}{EpisodeFileIDs: ids}, nil)
}

// BulkUpdateEpisodeFiles updates multiple episode files.
func (c *Client) BulkUpdateEpisodeFiles(ctx context.Context, editor *EpisodeFileEditorResource) ([]EpisodeFile, error) {
	var out []EpisodeFile
	if err := c.base.Put(ctx, "/api/v3/episodefile/bulk", editor, &out); err != nil {
		return nil, err
	}
	return out, nil
}
