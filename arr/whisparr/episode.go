package whisparr

import (
	"context"
	"fmt"

	"github.com/golusoris/goenvoy/arr/v2"
)

// GetEpisodes returns all episodes for a series.
func (c *Client) GetEpisodes(ctx context.Context, seriesID int) ([]Episode, error) {
	path := fmt.Sprintf("/api/v3/episode?seriesId=%d", seriesID)
	return arr.GetList[Episode](ctx, c.base, path)
}

// GetEpisode returns a single episode by ID.
func (c *Client) GetEpisode(ctx context.Context, id int) (*Episode, error) {
	path := fmt.Sprintf("/api/v3/episode/%d", id)
	return arr.GetOne[Episode](ctx, c.base, path)
}

// GetEpisodeFiles returns all episode files for a series.
func (c *Client) GetEpisodeFiles(ctx context.Context, seriesID int) ([]EpisodeFile, error) {
	path := fmt.Sprintf("/api/v3/episodefile?seriesId=%d", seriesID)
	return arr.GetList[EpisodeFile](ctx, c.base, path)
}

// DeleteEpisodeFile deletes an episode file by ID.
func (c *Client) DeleteEpisodeFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v3/episodefile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// UpdateEpisode updates an episode (e.g. monitored status).
func (c *Client) UpdateEpisode(ctx context.Context, ep *Episode) (*Episode, error) {
	return arr.PutOne[Episode](ctx, c.base, fmt.Sprintf("/api/v3/episode/%d", ep.ID), ep)
}

// MonitorEpisodes sets the monitored status for multiple episodes.
func (c *Client) MonitorEpisodes(ctx context.Context, req *EpisodesMonitoredResource) ([]Episode, error) {
	return arr.PutList[Episode](ctx, c.base, "/api/v3/episode/monitor", req)
}

// GetEpisodeFile returns a single episode file by ID.
func (c *Client) GetEpisodeFile(ctx context.Context, id int) (*EpisodeFile, error) {
	return arr.GetOne[EpisodeFile](ctx, c.base, fmt.Sprintf("/api/v3/episodefile/%d", id))
}

// UpdateEpisodeFile updates an episode file (quality, language).
func (c *Client) UpdateEpisodeFile(ctx context.Context, ef *EpisodeFile) (*EpisodeFile, error) {
	return arr.PutOne[EpisodeFile](ctx, c.base, fmt.Sprintf("/api/v3/episodefile/%d", ef.ID), ef)
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
	return arr.PutList[EpisodeFile](ctx, c.base, "/api/v3/episodefile/bulk", editor)
}
