package sonarr

import (
	"context"
	"fmt"
)

// GetEpisodes returns all episodes for the given series.
func (c *Client) GetEpisodes(ctx context.Context, seriesID int) ([]Episode, error) {
	var out []Episode
	path := fmt.Sprintf("/api/v3/episode?seriesId=%d", seriesID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEpisode returns a single episode by its database ID.
func (c *Client) GetEpisode(ctx context.Context, id int) (*Episode, error) {
	var out Episode
	path := fmt.Sprintf("/api/v3/episode/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEpisode updates the metadata for an episode (typically the monitored flag).
func (c *Client) UpdateEpisode(ctx context.Context, episode *Episode) (*Episode, error) {
	var out Episode
	path := fmt.Sprintf("/api/v3/episode/%d", episode.ID)
	if err := c.base.Put(ctx, path, episode, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MonitorEpisodes sets the monitored flag for a batch of episode IDs.
func (c *Client) MonitorEpisodes(ctx context.Context, episodeIDs []int, monitored bool) error {
	body := EpisodesMonitoredResource{EpisodeIDs: episodeIDs, Monitored: monitored}
	return c.base.Put(ctx, "/api/v3/episode/monitor", body, nil)
}

// GetEpisodeFiles returns all episode files for the given series.
func (c *Client) GetEpisodeFiles(ctx context.Context, seriesID int) ([]EpisodeFile, error) {
	var out []EpisodeFile
	path := fmt.Sprintf("/api/v3/episodefile?seriesId=%d", seriesID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEpisodeFile returns a single episode file by its database ID.
func (c *Client) GetEpisodeFile(ctx context.Context, id int) (*EpisodeFile, error) {
	var out EpisodeFile
	path := fmt.Sprintf("/api/v3/episodefile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEpisodeFile removes a single episode file by its database ID.
func (c *Client) DeleteEpisodeFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v3/episodefile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteEpisodeFiles removes multiple episode files by their IDs.
func (c *Client) DeleteEpisodeFiles(ctx context.Context, ids []int) error {
	body := EpisodeFileListResource{EpisodeFileIDs: ids}
	return c.base.Delete(ctx, "/api/v3/episodefile/bulk", &body, nil)
}

// UpdateEpisodeFile updates an individual episode file's metadata
// (quality, language, etc.).
func (c *Client) UpdateEpisodeFile(ctx context.Context, ef *EpisodeFile) (*EpisodeFile, error) {
	var out EpisodeFile
	path := fmt.Sprintf("/api/v3/episodefile/%d", ef.ID)
	if err := c.base.Put(ctx, path, ef, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditEpisodeFiles performs a bulk update of episode file metadata
// (quality, language, release group).
func (c *Client) EditEpisodeFiles(ctx context.Context, editor *EpisodeFileEditorResource) error {
	return c.base.Put(ctx, "/api/v3/episodefile/editor", editor, nil)
}

// ---------- Custom Format Bulk ----------.

// UpdateEpisodeFilesBulk performs a bulk update of episode file properties.
func (c *Client) UpdateEpisodeFilesBulk(ctx context.Context, editor *EpisodeFileEditorResource) ([]EpisodeFile, error) {
	var out []EpisodeFile
	if err := c.base.Put(ctx, "/api/v3/episodefile/bulk", editor, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Update Log File Content ----------.
